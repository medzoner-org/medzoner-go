---
slug: resilient-multi-client-esp32-camera
title: Building a resilient multi-client ESP32 camera
description: "How I designed an open-source ESP32 camera firmware for shared MJPEG streaming, adaptive load control, SD recording, and honest LAN security."
deck: "One small camera, several consumers, constrained memory and unreliable Wi-Fi. How I turned an ESP32-CAM streamer into an observable, adaptive and releasable open-source firmware."
author:
  name: Mehdi Youb
  url: https://www.medzoner.com/
category: iot
tags: [esp32, open-source, embedded]
language: en
image: /public/images/esp32cam-social.png
image_alt: An ESP32 camera shares one captured frame with browsers, home automation clients and SD recording.
status: published
published_at: 2026-10-24T16:14:59+02:00
---

## A camera endpoint was not enough {#why}

My initial requirement sounded small: make an ESP32-CAM feed available to Domoticz and the other tools on my home network. In practice, a browser, MotionEye, Home Assistant and OctoPrint may all ask for the same stream. A design that works for one tab can stall when a second slow client holds a frame or when weak Wi-Fi turns every socket write into back-pressure.

[domoticz-esp32cam](https://github.com/Medzoner/domoticz-esp32cam) is the firmware that grew from that constraint. It provides JPEG snapshots, one shared MJPEG stream for up to four clients, camera controls, health telemetry, first-boot Wi-Fi provisioning and optional SD recording. It runs on four PSRAM-equipped camera boards through PlatformIO.

> **Project status:** this is a personal open-source firmware, not a security camera product or a surveillance platform. Hardware support and recording behavior differ by board. Test the exact board, sensor, power supply and network you intend to use.

## Capture once, serve several clients {#shared-stream}

The camera uses two PSRAM frame buffers and latest-frame capture. A dedicated stream task obtains the current JPEG and broadcasts that same buffer to active sessions in sequence instead of asking the sensor to capture one copy per client.

The MJPEG response combines the multipart boundary and JPEG header into one write, then sends image bytes in 16 KiB blocks. Writes have a bounded deadline: a slow consumer can delay clients later in that broadcast cycle, but it cannot keep the camera mutex or stream task blocked indefinitely.

```text
Browser ───────────────┐
MotionEye ─────────────┼── http://camera.local:81/stream
Home Assistant ────────┤          ↑
OctoPrint ─────────────┘     shared JPEG frame
```

![The ESP32 camera captures one JPEG frame, applies an adaptive stream profile, then shares it with MJPEG clients and the SD recorder while exposing telemetry.](/public/images/esp32cam-flow.svg "Capture is shared. Sequential delivery is bounded by socket deadlines, and telemetry closes the feedback loop.")

## Prefer graceful degradation to a frozen stream {#adaptive}

The default is deliberately conservative: 10 FPS. Every two seconds, an adaptive controller evaluates active clients, average frame delivery time, Wi-Fi RSSI, free heap and socket failures. It selects one of five pressure profiles, down to an emergency 4 FPS mode with additional JPEG compression.

Degradation happens immediately when pressure rises. Recovery is slower: three healthy evaluation windows are required per level. That asymmetry avoids oscillating between profiles when the network sits near a threshold. The user's configured JPEG quality remains the baseline; adaptive compression is a temporary offset rather than a hidden configuration change.

This does not create bandwidth or memory. It gives the firmware a bounded way to shed work, and it makes that decision visible through `/health`: target FPS, actual FPS, delivery time, pressure and quality offset are all observable.

## Separate control, live video and playback {#services}

The firmware exposes three small HTTP services instead of routing every operation through one busy loop:

- **Port 80:** dashboard, snapshots, health, camera controls and recording commands.
- **Port 81:** long-lived live MJPEG sessions.
- **Port 82:** recorded-file playback and download.

A slow recording download should not delay a health request. A control mutation should not tear down existing stream consumers. This separation is intentionally simple—there is still one constrained device—but it keeps unrelated workloads from sharing the same request path.

## Make constrained behavior visible {#interface}

The embedded dashboard controls resolution, JPEG quality, brightness, contrast, orientation and flash. It also shows Wi-Fi strength, heap and PSRAM usage, stream sessions, frame delivery time, adaptive pressure, stack headroom and recording state. On mobile, the same interface collapses to one column.

The browser polls health every two seconds without overlapping requests and pauses when the tab is hidden. It also disconnects its own stream while hidden. Observability should not become the source of unnecessary work it is meant to explain.

![Responsive ESP32-CAM dashboard with a simulated camera feed, camera settings, system health and SD recording controls.](/public/images/esp32cam-dashboard.png "The published screenshot uses a simulated feed and demonstration telemetry; it contains no private camera image or network credential.")

## Record JPEGs without re-encoding them {#recording}

On the AI Thinker ESP32-CAM, the same captured JPEG frames can be written to the microSD card in a small MJR container. Segments are capped at 64 MiB. The integrated player reads those frames back as MJPEG at approximately 10 FPS, so recording does not require a video codec on the microcontroller.

The firmware reserves 8 MiB of free space and stops with an error rather than silently deleting older files. Recording continues even without a live viewer because capture is a device responsibility, not a side effect of a browser connection.

> **Hardware boundary:** SD recording is currently defined and validated for the AI Thinker board in 1-bit `SD_MMC` mode. Other supported boards expose streaming but report recording as unavailable until their storage wiring is explicitly supported.

## State the security boundary honestly {#security}

The firmware has optional HTTP Basic authentication, per-boot session cookies, same-origin checks for browser mutations, mutation rate limiting and bounded request parsing. Snapshots are copied out of the camera buffer before network delivery so a slow client cannot retain the camera resource.

It still has **no TLS**. Basic credentials, video and commands are not encrypted. There is no secure boot or flash encryption in the project configuration. The intended boundary is an isolated, trusted IoT network with firewall rules—not direct Internet exposure.

Calling this out is part of the design. Authentication can reduce accidental access on a trusted LAN, but it must not be presented as transport security.

## Keep hardware at the edges {#architecture}

The source tree borrows a ports-and-adapters shape that is useful even on a microcontroller. Platform-independent rules for HTTP parsing, camera controls, provisioning, recording and stream pressure live in `src/core`. Contracts live in `src/ports`. Arduino, camera, network, SD and HTTP details live in `src/adapters`. The `main.cpp` file is the composition root.

That boundary makes native Unity tests possible without an ESP32 attached. CI runs those tests, Cppcheck for native and ESP32 environments, and firmware builds for AI Thinker ESP32-CAM, ESP-EYE, ESP32-S3 CAM LCD and M5Stack Camera PSRAM.

```text
src/core       rules that can run natively
src/ports      contracts required by those rules
src/adapters   ESP32, camera, HTTP, Wi-Fi and SD details
src/main.cpp   explicit composition
```

## Ship a project, not a source dump {#open-source}

Publishing the repository also meant documenting contribution setup, security reporting and support boundaries; adding structured issue forms and CODEOWNERS; protecting `main`; and building every supported board in CI. Release Please derives Semantic Versions from Conventional Commits. A release attaches four board-specific binaries, an AI Thinker flashing bundle and SHA-256 checksums.

Those mechanics do not prove that every board works in every environment. They make changes reviewable and releases reproducible enough for another person to inspect, flash and report a precise result.

## What this design does not solve {#limits}

- Four clients is a deliberate bound, not a benchmark for arbitrary fan-out.
- MJPEG is simple and broadly compatible, but less bandwidth-efficient than modern video codecs.
- Adaptive profiles react to measured pressure; they cannot repair poor power, antennas or congested Wi-Fi.
- Only the AI Thinker storage path is currently supported for recording.
- The trusted-LAN model is unsuitable for direct exposure to an untrusted network.

The useful result is not that an ESP32-CAM becomes a full NVR. It is that a small device has explicit resource limits, observable degradation and enough separation to remain understandable.

## Flash it, measure it, improve it {#try}

Start with the AI Thinker environment or choose one of the other supported PSRAM boards. Use placeholder credentials for builds, keep real credentials in the ignored local header, and test stream behavior with the clients you actually run.

> [!CTA]
>
> **Want to inspect the firmware?**
>
> Read the architecture, download a checked release, or contribute a focused hardware result.
>
> [Explore domoticz-esp32cam on GitHub →](https://github.com/Medzoner/domoticz-esp32cam)
>
> [Download a release](https://github.com/Medzoner/domoticz-esp32cam/releases) · [Report an issue](https://github.com/Medzoner/domoticz-esp32cam/issues) · [Contribute](https://github.com/Medzoner/domoticz-esp32cam/blob/main/CONTRIBUTING.md)

I maintain this project and welcome independent tests, especially on the supported non-AI-Thinker boards. Remove credentials, private addresses and camera images from reports.
