local album_slug_key = KEYS[1]
local entity_prefix = ARGV[1]

local album_id = redis.call("GET", album_slug_key)
if not album_id then
    return {err = "album_not_found"}
end

local album_entity_key = entity_prefix .. album_id
local album_json = redis.call("GET", album_entity_key)

return album_json