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

#------------------------------------------------------------------------------------------
# POST "CONTACT" - Test failed
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - POST] "Home page - Test failed"
        And I add "Content-Type" header equal to "application/x-www-form-urlencoded"
        When    I send a POST request to "/" with body:
          """
          {"name": "", "email": "", "message": ""}
          """
        Then    the response status code should be 400
