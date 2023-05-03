package adguard

// adguard_config defaults
const CONFIG_FILTERING_ENABLED = true
const CONFIG_FILTERING_UPDATE_INTERVAL uint = 24 // hours
const CONFIG_SAFEBROWSING_ENABLED = false
const CONFIG_PARENTAL_CONTROL_ENABLED = false
const CONFIG_SAFE_SEARCH_ENABLED = false
const CONFIG_QUERYLOG_ENABLED = true
const CONFIG_QUERYLOG_INTERVAL uint64 = 2160 // hours
const CONFIG_QUERYLOG_ANONYMIZE_CLIENT_IP = false
const CONFIG_STATS_ENABLED = true
const CONFIG_STATS_INTERVAL = 24 // hours
const CONFIG_DNS_RATE_LIMIT = 20
const CONFIG_DNS_BLOCKING_MODE = "default"
const CONFIG_DNS_EDNS_CS_ENABLED = false
const CONFIG_DNS_DISABLE_IPV6 = false
const CONFIG_DNS_DNSSEC_ENABLED = false
const CONFIG_DNS_CACHE_SIZE = 4194304
const CONFIG_DNS_CACHE_TTL_MIN = 0
const CONFIG_DNS_CACHE_TTL_MAX = 0
const CONFIG_DNS_CACHE_OPTIMISTIC = false
const CONFIG_DNS_UPSTREAM_MODE = "load_balance"
const CONFIG_DNS_USE_PRIVATE_PTR_RESOLVERS = true
const CONFIG_DNS_RESOLVE_CLIENTS = true
const CONFIG_DHCP_ENABLED = false
const CONFIG_DHCP_LEASE_DURATION = 86400 // seconds

var CONFIG_DNS_BOOTSTRAP = []string{"9.9.9.10", "149.112.112.10", "2620:fe::10", "2620:fe::fe:10"}
var CONFIG_DNS_UPSTREAM = []string{"https://dns10.quad9.net/dns-query"}
var CONFIG_SAFE_SEARCH_SERVICES_OPTIONS = []string{"bing", "duckduckgo", "google", "pixabay", "yandex", "youtube"}
var CONFIG_GLOBAL_BLOCKED_SERVICES_OPTIONS = []string{"9gag", "amazon", "bilibili", "cloudflare", "crunchyroll", "dailymotion", "deezer",
	"discord", "disneyplus", "douban", "ebay", "epic_games", "facebook", "gog", "hbomax", "hulu", "icloud_private_relay", "imgur",
	"instagram", "iqiyi", "kakaotalk", "lazada", "leagueoflegends", "line", "mail_ru", "mastodon", "minecraft", "netflix", "ok",
	"onlyfans", "origin", "pinterest", "playstation", "qq", "rakuten_viki", "reddit", "riot_games", "roblox", "shopee", "skype", "snapchat",
	"soundcloud", "spotify", "steam", "telegram", "tiktok", "tinder", "twitch", "twitter", "valorant", "viber", "vimeo", "vk", "voot", "wechat",
	"weibo", "whatsapp", "xboxlive", "youtube", "zhihu"}
