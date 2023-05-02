package adguard

// adguard_config
const FILTERING_ENABLED = true
const FILTERING_UPDATE_INTERVAL uint = 24 // hours
const SAFEBROWSING_ENABLED = false
const PARENTAL_CONTROL_ENABLED = false
const SAFE_SEARCH_ENABLED = false
var SAFE_SEARCH_SERVICES = []string{"bing", "duckduckgo", "google", "pixabay", "yandex", "youtube"}
const QUERYLOG_ENABLED = true
const QUERYLOG_INTERVAL uint64 = 2160  // hours
const QUERYLOG_ANONYMIZE_CLIENT_IP = false
const STATS_ENABLED = true
const STATS_INTERVAL = 24 // hours
var BLOCKED_SERVICES_ALL = []string{"9gag", "amazon", "bilibili", "cloudflare", "crunchyroll", "dailymotion", "deezer",
	"discord", "disneyplus", "douban", "ebay", "epic_games", "facebook", "gog", "hbomax", "hulu", "icloud_private_relay", "imgur",
	"instagram", "iqiyi", "kakaotalk", "lazada", "leagueoflegends", "line", "mail_ru", "mastodon", "minecraft", "netflix", "ok",
	"onlyfans", "origin", "pinterest", "playstation", "qq", "rakuten_viki", "reddit", "riot_games", "roblox", "shopee", "skype", "snapchat",
	"soundcloud", "spotify", "steam", "telegram", "tiktok", "tinder", "twitch", "twitter", "valorant", "viber", "vimeo", "vk", "voot", "wechat",
	"weibo", "whatsapp", "xboxlive", "youtube", "zhihu"}
var DNS_BOOTSTRAP = []string{"9.9.9.10", "149.112.112.10", "2620:fe::10", "2620:fe::fe:10"}
var DNS_UPSTREAM = []string{"https://dns10.quad9.net/dns-query"}
const DNS_RATE_LIMIT = 20
const DNS_BLOCKING_MODE = "default"
const DNS_EDNS_CS_ENABLED = false