Feature: Static assets and health probes
    In order to check the infrastructure endpoints
    As a monitoring system
    I need to access health probes and static files

    Background:
        And I add "Authorization" header equal to ""

#------------------------------------------------------------------------------------------
# GET "Liveness probe"
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET] "Liveness probe"
        When    I send a GET request to "/healthz/live"
        Then    the response status code should be 200

#------------------------------------------------------------------------------------------
# GET "Readiness probe"
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET] "Readiness probe"
        When    I send a GET request to "/healthz/ready"
        Then    the response status code should be 200

#------------------------------------------------------------------------------------------
# GET "Static asset - robots.txt"
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET] "Static robots.txt"
        When    I send a GET request to "/public/robots.txt"
        Then    the response status code should be 200

#------------------------------------------------------------------------------------------
# GET "Static asset - CSS"
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET] "Static CSS"
        When    I send a GET request to "/public/css/app.css"
        Then    the response status code should be 200

