-- Query over the a range of users and their active segmentations.
SELECT user_id,
    JSON_AGG(arr->'segment'->'name')
FROM segment_profiles,
    json_array_elements(segments) AS arr
WHERE (arr->>'member')::BOOLEAN
    AND user_id < 'usr-000fff' -- Fetch only the first chunk of the total range.
GROUP BY user_id;