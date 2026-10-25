---
slug: resilient-multi-client-esp32-camera
title: Building a resilient multi-client ESP32 camera
description: "How an ESP32 camera shares bounded JPEG buffers, recovers from interrupted streams, handles SD contention and makes its reliability limits explicit."
deck: "One small camera, several consumers, constrained memory and unreliable Wi-Fi. What changed when native tests met partial JPEGs, silent browser disconnects and real hardware."
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
updated_at: 2026-10-25T19:14:55+01:00
---

## A camera endpoint was not enough {#why}

My initial requirement sounded small: make an ESP32-CAM feed available to Domoticz and the other tools on my home network. In practice, a browser, MotionEye, Home Assistant and OctoPrint may all ask for the same stream. A design that works for one tab can stall when a second slow client holds a frame or when weak Wi-Fi turns every socket write into back-pressure.

[domoticz-esp32cam](https://github.com/Medzoner/domoticz-esp32cam) is the firmware that grew from that constraint. It provides JPEG snapshots, one shared MJPEG stream for up to four clients, camera controls, health telemetry, first-boot Wi-Fi provisioning and optional SD recording. It runs on four PSRAM-equipped camera boards through PlatformIO.

> **Project status:** this is a personal open-source firmware, not a security camera product or a surveillance platform. Hardware support and recording behavior differ by board. Test the exact board, sensor, power supply and network you intend to use.

> **Reviewed implementation:** this update follows [firmware commit 96cb9ae](https://github.com/Medzoner/domoticz-esp32cam/tree/96cb9ae32c83d5875782a9384346b22f426aa497). It includes changes merged after the original article; release assets may lag `main`. The versioned [reliability contract](https://github.com/Medzoner/domoticz-esp32cam/blob/96cb9ae32c83d5875782a9384346b22f426aa497/docs/RELIABILITY.md) distinguishes automated checks from physical hardware evidence.

## Capture once, serve several clients {#shared-stream}

The camera uses two PSRAM driver buffers and latest-frame capture. A dedicated stream task copies the current JPEG into one reusable 512 KiB PSRAM delivery buffer, returns the driver's frame, and releases the camera mutex **before** network or SD writes. All live consumers receive that same copy in sequence instead of asking the sensor to capture one frame per client.

That extra copy is an explicit trade-off: 512 KiB reserved once, no allocation in the frame loop, and no fallback that could exhaust internal RAM. A failed reservation makes stream startup fail visibly. An oversized JPEG is dropped whole rather than truncated or used to grow memory. `/health` exposes `stream_frame_capacity`, `stream_dropped_frames` and `stream_oversized_frames`. Snapshots have a separate allocation path.

The driver allocation must also match the resolutions exposed by the UI. Following Espressif's CameraWebServer example, the firmware initializes at UXGA/high JPEG quality when PSRAM is present, then reduces the active sensor to VGA/default quality. Initializing at VGA and subsequently requesting larger frames risks exceeding the driver's original buffers. Without PSRAM, frame-size controls cannot exceed the VGA allocation. Missing JPEG start/end markers are rejected, but marker checks are not a JPEG decoder or a checksum.

The MJPEG response combines the multipart boundary and JPEG header, then sends bytes in blocks of at most 16 KiB. Using `send(..., MSG_DONTWAIT)` avoids the hidden retry loop inside `WiFiClient::write()`. A live JPEG gets a 300 ms no-progress deadline and a 2-second absolute budget. Progress refreshes only the idle deadline; it cannot extend the absolute cap. Headers and playback/download blocks retain their 300 ms total budgets.

A progressing high-resolution transfer should not be cut at the old 300 ms total deadline. But a slow consumer can still delay later clients: four progressing clients can use up to eight seconds of payload budgets, plus headers, capture, SD and scheduling. Releasing the camera mutex removes network/storage delivery from that critical section; it does not make delivery parallel or bound an SD driver call. `stream_write_failures` counts failed live writes, including disconnects, separately from rejected captures.

![Browser, MotionEye, Home Assistant and OctoPrint connect to the same ESP32-CAM MJPEG stream.](/public/images/esp32cam-clients.svg "Four clients connect to the same endpoint; capture remains shared.")

All four clients request `http://camera.local:81/stream`, backed by the shared JPEG frame.

![The ESP32 camera copies a JPEG into a bounded PSRAM delivery buffer, returns the driver frame, then branches to MJPEG clients and the SD recorder; a dashed line represents telemetry.](/public/images/esp32cam-flow.svg "Solid arrows carry the copied JPEG to up to four clients and optional SD recording; the dashed arrow represents telemetry, not image delivery.")

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

Shared resources still matter: health telemetry can wait for the SD mutex. A successful HTTP response must not turn that wait into a fictitious idle state. The recorder now returns an explicitly invalid snapshot after a failed lock instead of reading live 64-bit counters without ownership. `/health` reports `recording_stats_valid: false` and degraded status; clients must treat the associated recording fields as unknown.

Task creation is not readiness either. Video and playback move through `not_started`, `starting`, `running` and `failed`; only a listening socket makes a service running. A failed video/playback startup leaves HTTP controls available for diagnosis if their own server can start, instead of rebooting repeatedly. The diagnostic `/health` response remains HTTP 200 when JSON can be served: readiness monitors must inspect its `status` and service fields.

## Make constrained behavior visible {#interface}

The embedded dashboard controls resolution, JPEG quality, brightness, contrast, orientation and flash. It also shows Wi-Fi strength, heap and PSRAM usage, stream sessions, frame delivery time, adaptive pressure, stack headroom and recording state. On mobile, the same interface collapses to one column.

The browser polls health every two seconds without overlapping requests and pauses when the tab is hidden. It also disconnects its own stream while hidden. Observability should not become the source of unnecessary work it is meant to explain.

When recording telemetry is unknown—or the health request itself fails—the dashboard disables recording actions and clears stale metrics. Stop/list requests return HTTP 503 if their work cannot be accepted. An accepted stop is queued, not a claim that flushing has finished. Start and stop are serialized under the mutex: the last accepted command wins, and a failed command changes neither intent.

![Responsive ESP32-CAM dashboard with a simulated camera feed, camera settings, system health and SD recording controls.](/public/images/esp32cam-dashboard.png "The published screenshot uses a simulated feed and demonstration telemetry; it contains no private camera image or network credential.")

## A white band was a recovery bug, too {#browser-recovery}

A reported white band initially looked like a damaged camera image. The investigation found two different failure modes: undersized driver allocation, and network writes cut off while a large JPEG was still progressing. Neither explains every possible visual artifact, and decoding a set of complete JPEGs cannot prove that a user's intermittent symptom has disappeared.

The browser added another trap. In a local Firefox test, I cut a synthetic MJPEG response halfway through a JPEG. Firefox displayed an incomplete image and emitted no `error` event. The old code waited for that event to reconnect, so the stream remained frozen even though the camera's health endpoint was reachable.

The dashboard now preserves the user's stream intent and watches observed frame-load activity. Where the browser reports loads per MJPEG frame, roughly ten seconds without a new frame masks the stale image and schedules a new connection. A first image that never arrives has its own timeout. Health fetch and body consumption have a four-second abort budget; offline/online events, camera reboot and service-readiness changes also help recovery. Retries coalesce and back off from one to ten seconds.

Each attempt uses a fresh image node with generation-bound callbacks. A delayed event from the previous request cannot mark the new connection alive or tear it down. Stop, still capture and playback cancel live intent and pending timers. Reconnection must not undo the user's Stop button.

The before/after Firefox test reproduced the white lower region with a synthetic scene, then verified that the new code masked the stale image and resumed complete frames without relying on `onerror`. That is evidence for browser recovery, not a physical Wi-Fi outage test. Browsers without per-frame load events use health/network fallbacks, which cannot identify every dead individual session while other clients remain active. A raw `/stream` link has no recovery JavaScript; its client must reconnect itself. A partial image may still be briefly visible before detection.

## Record JPEGs without re-encoding them {#recording}

On the AI Thinker ESP32-CAM, the same copied JPEG frames can be written to the microSD card in a small MJR1 container. Segments are capped at 64 MiB. Playback is paced toward 10 FPS when I/O permits; that is not a throughput guarantee. Recording does not require a video codec on the microcontroller.

The firmware reserves 8 MiB of free space and stops with an error rather than silently deleting older files. Recording continues even without a live viewer because capture is a device responsibility, not a side effect of a browser connection.

Short header or payload writes fail the append and close the segment without counting an incomplete record as a finished frame. The player checks a frame's declared size against remaining file bytes before announcing its multipart header. A known truncated tail is therefore not advertised as another image; earlier structurally complete frames can still be played. Positive short reads are accumulated, while a premature zero read during download is an error rather than a successful EOF.

For an offline copy, `scripts/mjr_recover.py` can inspect the container or copy its structurally complete prefix to a **new** file. It never overwrites the original or an existing output path. MJR1 has no checksum, commit marker or expected final frame count: recovery cannot repair FAT, reconstruct missing bytes or prove JPEG validity. A later physical read failure can still interrupt a frame whose header was already sent.

Reader metadata is copied under the storage mutex. If closing a reader times out, playback closes its network client and retries cleanup before accepting another session instead of abandoning a reader that would block future recording. If the driver never recovers, cleanup can remain pending; this is not a whole-operation deadline. A full-length write and the SDK's `flush()` call are not durability guarantees.

> **Hardware boundary:** storage support targets the AI Thinker board in 1-bit `SD_MMC` mode. Other supported boards report recording as unavailable until storage wiring is implemented. The latest tests described here used filesystem/FreeRTOS doubles; the physical board had no available SD card. Do not present that as power-loss or card-failure validation.

## State the security boundary honestly {#security}

The firmware has optional HTTP Basic authentication, per-boot session cookies, same-origin checks for browser mutations, mutation rate limiting and bounded request parsing. Snapshots are copied out of the camera buffer before network delivery so a slow client cannot retain the camera resource.

It still has **no TLS**. Basic credentials, video and commands are not encrypted. There is no secure boot or flash encryption in the project configuration. The intended boundary is an isolated, trusted IoT network with firewall rules—not direct Internet exposure.

Calling this out is part of the design. Authentication can reduce accidental access on a trusted LAN, but it must not be presented as transport security.

## Keep hardware at the edges {#architecture}

The source tree borrows a ports-and-adapters shape that is useful even on a microcontroller. Platform-independent rules for HTTP parsing, camera controls, provisioning, recording and stream pressure live in `src/core`. Contracts live in `src/ports`. Arduino, camera, network, SD and HTTP details live in `src/adapters`. The `main.cpp` file is the composition root.

That boundary makes native Unity tests possible without an ESP32 attached. CI runs those tests, Cppcheck for native and ESP32 environments, and firmware builds for AI Thinker ESP32-CAM, ESP-EYE, ESP32-S3 CAM LCD and M5Stack Camera PSRAM.

Native suites also run with AddressSanitizer and UndefinedBehaviorSanitizer. Dedicated host environments compile the real SD adapter against instrumented FreeRTOS/filesystem doubles to force lock failures and check ownership, command ordering and retryable reader cleanup. Python tests exercise offline recovery; JavaScript tests use a fake clock/DOM/fetch for stream recovery. These are focused fault-injection tests, not a real scheduler, SD card, browser compositor or proof that every race is absent.

```text
src/core       rules that can run natively
src/ports      contracts required by those rules
src/adapters   ESP32, camera, HTTP, Wi-Fi and SD details
src/main.cpp   explicit composition
```

## Ship a project, not a source dump {#open-source}

Publishing the repository also meant documenting contribution setup, security reporting and support boundaries; adding structured issue forms and CODEOWNERS; protecting `main`; and building every supported board in CI. Release Please derives Semantic Versions from Conventional Commits. A release attaches four board-specific binaries, an AI Thinker flashing bundle and SHA-256 checksums.

Those mechanics do not prove that every board works in every environment. They make changes reviewable and releases reproducible enough for another person to inspect, flash and report a precise result.

This iteration was also flashed to a physical AI Thinker, with esptool hash verification and HTTP-only startup checks. Short JPEG sampling checked decoding and dimensions in VGA, XGA and UXGA; a high-resolution transfer cutoff was reproduced before the deadline correction, followed by a successful run of complete UXGA frames afterward. Runtime controls were restored, and private camera frames were neither saved nor published. That is a hardware smoke test—not a 24-hour soak, validation of the other three boards, SD testing or confirmation of recovery from a real AP outage.

## What this design does not solve {#limits}

- Four clients is a deliberate bound, not a benchmark for arbitrary fan-out.
- MJPEG is simple and broadly compatible, but less bandwidth-efficient than modern video codecs.
- Adaptive profiles react to measured pressure; they cannot repair poor power, antennas or congested Wi-Fi.
- Only the AI Thinker storage path is currently supported for recording.
- The fixed delivery copy deliberately drops JPEGs above 512 KiB; larger driver buffers do not remove that cap.
- Sequential delivery and SD contention can reduce actual FPS far below the configured target.
- Browser recovery can hide and reconnect after a partial image, not guarantee that one is never displayed.
- The trusted-LAN model is unsuitable for direct exposure to an untrusted network.

The useful result is not that an ESP32-CAM becomes a full NVR. It is that a small device has explicit resource limits, observable degradation and enough separation to remain understandable.

## Keep the article accountable to the code {#evolution}

The commit linked near the introduction is the reviewed implementation, not a promise that this page always describes the latest `main`. The site repository tracks that baseline and watches relevant upstream files. When they change, a scheduled check opens or updates one editorial-review issue with a comparison link; it does not silently rewrite claims, advance the baseline or publish an unreviewed article.

An update should reconcile source contracts, diagrams and tests, distinguish new hardware evidence from build coverage, and only then advance the reviewed commit and article update date through the site's normal pull-request workflow. A drift notification is a request to investigate, not proof that the article is wrong—or that a new feature has been validated.

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
