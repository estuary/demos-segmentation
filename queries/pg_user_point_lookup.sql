-- Point lookup of active and inactive segmentations for a specific user.
SELECT user_id,
    arr->'segment'->>'vendor' AS vendor,
    arr->'segment'->>'name' AS segment,
    arr->'member' AS current_member,
    (arr->>'first')::TIMESTAMP AS first_seen,
    (arr->>'last')::TIMESTAMP AS last_update
FROM segment_profiles,
    json_array_elements(segments) as arr
WHERE user_id = 'usr-0007db';