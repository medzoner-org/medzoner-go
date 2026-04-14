Feature: Global behavior "Medzoner"
    In order to check the global behavior APP
    As a visitor
    I need to able to access

    Background:
        And I add "Authorization" header equal to ""

#------------------------------------------------------------------------------------------
# GET "Home" - Test succeeded
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET_ALL] "Home page"
        When    I send a GET request to "/"
        Then    the response status code should be 200
        And     the response body should contain "MedZoner.com"

#------------------------------------------------------------------------------------------
# GET "Home" - With TOR-HOST header
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET] "Home page with TOR-HOST"
        And I add "TOR-HOST" header equal to "test-onion-host"
        When    I send a GET request to "/"
        Then    the response status code should be 200
        And     the response body should contain "MedZoner.com"

#------------------------------------------------------------------------------------------
# POST "CONTACT" - Test success
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - POST] "Home page - Test success"
        And I add "Content-Type" header equal to "application/x-www-form-urlencoded"
        When    I send a POST request to "/" with body:
          """
          {"name": "else", "email": "email@fake.com", "message": "else"}
          """
        Then    the response status code should be 303
        And     the response header "Location" should contain "/#contact"

#------------------------------------------------------------------------------------------
# POST "CONTACT" - Test failed validation
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - POST] "Home page - Test failed"
        And I add "Content-Type" header equal to "application/x-www-form-urlencoded"
        When    I send a POST request to "/" with body:
          """
          {"name": "", "email": "", "message": ""}
          """
        Then    the response status code should be 400
        And     the response body should contain "Contact me"

#------------------------------------------------------------------------------------------
# POST "CONTACT" - With submit button (renders form, no processing)
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - POST] "Home page - With submit param"
        And I add "Content-Type" header equal to "application/x-www-form-urlencoded"
        When    I send a POST request to "/" with body:
          """
          {"name": "test", "email": "test@test.com", "message": "msg", "submit": "Send"}
          """
        Then    the response status code should be 200
        And     the response body should contain "MedZoner.com"

