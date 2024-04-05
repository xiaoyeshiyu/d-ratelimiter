local key = KEYS[1]

local start = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[4])

redis.call('ZREMRANGEBYSCORE', key, '-inf', start)

local count = redis.call('ZCOUNT', key, '-inf', '+inf')

if count >= threshold then
    return "false"
else
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return "true"
end