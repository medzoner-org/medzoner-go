Feature: Global behavior "Medzoner"
    In order to check the global behavior
    As a visitor
    I need to able to access APP

    Background:
        And I add "Authorization" header equal to ""

    Scenario: [Medzoner - GET - success] "404 page"
        When    I send a GET request to "/not-found"
        Then    the response status code should be 404

    Scenario: [Medzoner - GET - success] "Another 404 page"
        When    I send a GET request to "/does-not-exist"
        Then    the response status code should be 404

