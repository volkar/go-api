local user_entity_key = KEYS[1]
local mapper_prefix = ARGV[1]

local user_json = redis.call("GET", user_entity_key)
if not user_json then
    return 0
end

local ok, user_obj = pcall(cjson.decode, user_json)

if ok and type(user_obj) == "table" and user_obj.slug then
    local mapper_key = mapper_prefix .. user_obj.slug
    redis.call("UNLINK", mapper_key, user_entity_key)
    return 2
end

redis.call("UNLINK", user_entity_key)
return 1