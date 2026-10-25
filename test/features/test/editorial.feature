Feature: Markdown editorial catalogue
  As a reader
  I want to discover the published catalogue and subscribe

  Scenario: Browse published articles
    When I send a GET request to "/articles"
    Then the response status code should be 200
    And the response body should contain "/articles/after-wire-whyor"
    And the response body should contain "/articles/resilient-multi-client-esp32-camera"

  Scenario: Filter a rubric
    When I send a GET request to "/articles?category=go"
    Then the response status code should be 200
    And the response body should contain "After Wire: compile-time dependency injection with whyor"

  Scenario: Discover publications in RSS
    When I send a GET request to "/feed.xml"
    Then the response status code should be 200
    And the response body should contain "<rss version="
    And the response body should contain "https://www.medzoner.com/articles/after-wire-whyor"

  Scenario: Discover the dynamic sitemap
    When I send a GET request to "/sitemap.xml"
    Then the response status code should be 200
    And the response body should contain "https://www.medzoner.com/articles"

  Scenario: Reject invalid pagination
    When I send a GET request to "/articles?page=-1"
    Then the response status code should be 400
