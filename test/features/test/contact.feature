Feature: Global behavior "Medzoner"
    In order to check the contact behavior APP
    As a visitor
    I need to able to access

    Background:
        And I add "Content-Type" header equal to "text/html"

#------------------------------------------------------------------------------------------
# GET "CONTACT" - Test succeeded
#------------------------------------------------------------------------------------------

    Scenario: [Medzoner - GET_ALL] "Contact page"
        When    I send a GET request to "/contact"
        Then    the response status code should be 404
