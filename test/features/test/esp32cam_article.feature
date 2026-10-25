Feature: ESP32-CAM project article
  As an embedded developer
  I want to read how the multi-client camera firmware was designed
  So I can understand its trade-offs and evaluate the project

  Scenario: Read the public article
    When I send a GET request to "/articles/resilient-multi-client-esp32-camera"
    Then the response status code should be 200
    And the response body should contain "Building a resilient multi-client ESP32 camera"
    And the response body should contain "http://camera.local:81/stream"
    And the response body should contain "https://github.com/Medzoner/domoticz-esp32cam"
    And the response body should contain "https://www.medzoner.com/articles/resilient-multi-client-esp32-camera"

  Scenario: Discover the article from the home page
    When I send a GET request to "/"
    Then the response status code should be 200
    And the response body should contain "/articles/resilient-multi-client-esp32-camera"
    And the response body should contain "ESP32-CAM : streaming vidéo multi-clients"

  Scenario: Discover the article from the sitemap
    When I send a GET request to "/sitemap.xml"
    Then the response status code should be 200
    And the response body should contain "https://www.medzoner.com/articles/resilient-multi-client-esp32-camera"
