-- Rename saved views from the old application-landscape type to capability-landscape
-- (the view was renamed to reflect that it shows capabilities, not domains).
UPDATE saved_views
SET view_type = 'capability-landscape'
WHERE view_type = 'application-landscape';
