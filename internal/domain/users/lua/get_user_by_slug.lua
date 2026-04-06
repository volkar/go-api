local user_slug_key = KEYS[1]
local entity_prefix = ARGV[1]

local user_id = redis.call("GET", user_slug_key)
if not user_id then
    return {err = "user_not_found"}
end

local user_entity_key = entity_prefix .. user_id
local user_json = redis.call("GET", user_entity_key)

return user_json