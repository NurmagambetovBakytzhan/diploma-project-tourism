WITH empty_check AS (
    SELECT COUNT(*) = 0 AS is_empty FROM tourism.categories
)
INSERT INTO tourism.categories (name)
SELECT * FROM (VALUES
                   ('Automotive'),
                   ('Business'),
                   ('Culture'),
                   ('Education'),
                   ('Entertainment and Recreation'),
                   ('Facilities'),
                   ('Finance'),
                   ('Food and Drink'),
                   ('Geographical Areas'),
                   ('Government'),
                   ('Health and Wellness'),
                   ('Housing'),
                   ('Lodging'),
                   ('Natural Features'),
                   ('Places of Worship'),
                   ('Services'),
                   ('Shopping'),
                   ('Sports'),
                   ('Transportation')
              ) AS new_categories(name)
WHERE (SELECT is_empty FROM empty_check);