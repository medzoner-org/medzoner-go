Feature: whyor launch article
  As a Go developer
  I want to read the whyor launch article
  So I can understand its generated code and migration limits

  Scenario: Read the public article
    When I send a GET request to "/articles/after-wire-whyor"
    Then the response status code should be 200
    And the response body should contain "After Wire: compile-time dependency injection with whyor"
    And the response body should contain "whyor.Bind[Store, *MemoryStore]()"
    And the response body should contain "https://github.com/Medzoner/whyor"
    And the response body should contain "https://www.medzoner.com/articles/after-wire-whyor"

  Scenario: Discover the article from the home page
    When I send a GET request to "/"
    Then the response status code should be 200
    And the response body should contain "/articles/after-wire-whyor"
    And the response body should contain "After Wire, meet whyor."

  Scenario: Unknown article remains a 404
    When I send a GET request to "/articles/not-an-article"
    Then the response status code should be 404
