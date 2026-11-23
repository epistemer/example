// Code generated from labels by github.com/epistemer/i18n. DO NOT EDIT.

package settings

import "github.com/epistemer/i18n"

var (
	L10nLog                       = i18n.Key("settings.log", "Logging")
	L10nLogLevel                  = i18n.Key("settings.log.level", "Minimum log level")
	L10nLogFormat                 = i18n.Key("settings.log.format", "Output format")
	L10nServer                    = i18n.Key("settings.server", "HTTP server")
	L10nServerAddr                = i18n.Key("settings.server.addr", "Address to listen on")
	L10nServerPublicURL           = i18n.Key("settings.server.public_url", "Address people reach the app at")
	L10nServerShutdownTimeout     = i18n.Key("settings.server.shutdown_timeout", "How long to wait for requests to finish when stopping")
	L10nStorage                   = i18n.Key("settings.storage", "Storage")
	L10nStorageBolt               = i18n.Key("settings.storage.bolt", "Store file: settings, translations and sign-ins")
	L10nStorageFiles              = i18n.Key("settings.storage.files", "Directory of uploaded files")
	L10nDB                        = i18n.Key("settings.db", "Database")
	L10nDBURL                     = i18n.Key("settings.db.url", "PostgreSQL connection URL")
	L10nMail                      = i18n.Key("settings.mail", "Outgoing mail")
	L10nMailOutbox                = i18n.Key("settings.mail.outbox", "Directory each message is written to, one file per message")
	L10nAuth                      = i18n.Key("settings.auth", "Sign-in and sessions")
	L10nAuthIssuer                = i18n.Key("settings.auth.issuer", "Issuer: the name in tokens and authenticator apps")
	L10nAuthSecret                = i18n.Key("settings.auth.secret", "Master secret, at least 32 bytes; every key is derived from it")
	L10nAuthSecretPrevious        = i18n.Key("settings.auth.secret_previous", "Previous master secret, while rotating")
	L10nAuthTrustedOrigins        = i18n.Key("settings.auth.trusted_origins", "Origins whose cross-origin requests are trusted")
	L10nAuthSession               = i18n.Key("settings.auth.session", "Sessions")
	L10nAuthSessionIdle           = i18n.Key("settings.auth.session.idle", "Idle lifetime of a session")
	L10nAuthSessionRememberedIdle = i18n.Key("settings.auth.session.remembered_idle", "Idle lifetime of a remembered session")
	L10nAuthSessionLifetime       = i18n.Key("settings.auth.session.lifetime", "Longest lifetime of a session, whatever its use")
	L10nAuthAccess                = i18n.Key("settings.auth.access", "Access tokens")
	L10nAuthAccessLifetime        = i18n.Key("settings.auth.access.lifetime", "Lifetime of an access token")
	L10nAuthRefresh               = i18n.Key("settings.auth.refresh", "Refresh tokens")
	L10nAuthRefreshGrace          = i18n.Key("settings.auth.refresh.grace", "How long a replaced refresh token still returns its successor")
	L10nAuthFlow                  = i18n.Key("settings.auth.flow", "Sign-in flows")
	L10nAuthFlowLifetime          = i18n.Key("settings.auth.flow.lifetime", "Lifetime of a sign-in flow")
	L10nAuthFresh                 = i18n.Key("settings.auth.fresh", "Fresh proofs")
	L10nAuthFreshWindow           = i18n.Key("settings.auth.fresh.window", "How recent a proof must be to change a factor")
	L10nAuthCode                  = i18n.Key("settings.auth.code", "Codes sent by email or SMS")
	L10nAuthCodeLength            = i18n.Key("settings.auth.code.length", "Digits in a code")
	L10nAuthCodeLifetime          = i18n.Key("settings.auth.code.lifetime", "Lifetime of a code")
	L10nAuthCodeTries             = i18n.Key("settings.auth.code.tries", "Wrong tries a code allows")
	L10nAuthCodeResend            = i18n.Key("settings.auth.code.resend", "Wait before another code may be sent")
	L10nAuthCodeHourly            = i18n.Key("settings.auth.code.hourly", "Codes a user may be sent in an hour")
	L10nAuthLink                  = i18n.Key("settings.auth.link", "Links sent by email")
	L10nAuthLinkLifetime          = i18n.Key("settings.auth.link.lifetime", "Lifetime of a recovery link")
	L10nAuthInvitation            = i18n.Key("settings.auth.invitation", "Invitations")
	L10nAuthInvitationLifetime    = i18n.Key("settings.auth.invitation.lifetime", "Lifetime of an invitation")
	L10nAuthDevice                = i18n.Key("settings.auth.device", "Registered devices")
	L10nAuthDeviceLifetime        = i18n.Key("settings.auth.device.lifetime", "Lifetime of a registered device")
	L10nAuthPassword              = i18n.Key("settings.auth.password", "Passwords")
	L10nAuthPasswordMinLength     = i18n.Key("settings.auth.password.min_length", "Fewest characters in a password")
	L10nAuthPasswordMinScore      = i18n.Key("settings.auth.password.min_score", "Least strength a new password needs, 3 or 4")
	L10nAuthPasswordMemory        = i18n.Key("settings.auth.password.memory", "Memory for each password hash, in MiB")
	L10nAuthPasswordTime          = i18n.Key("settings.auth.password.time", "Passes of each password hash")
	L10nAuthPasswordConcurrency   = i18n.Key("settings.auth.password.concurrency", "Password hashes at once")
	L10nAuthThrottle              = i18n.Key("settings.auth.throttle", "Throttling of wrong proofs")
	L10nAuthThrottleFree          = i18n.Key("settings.auth.throttle.free", "Failures before waits start")
	L10nAuthThrottleCeiling       = i18n.Key("settings.auth.throttle.ceiling", "Longest wait")
	L10nAuthThrottleWindow        = i18n.Key("settings.auth.throttle.window", "Time without a failure after which a counter starts again")
	L10nAuthCookie                = i18n.Key("settings.auth.cookie", "Cookies")
	L10nAuthCookieInsecure        = i18n.Key("settings.auth.cookie.insecure", "Drop Secure and the __Host- prefix, for development over plain HTTP")
	L10nTheme                     = i18n.Key("settings.theme", "Theme")
	L10nThemeBrandColor           = i18n.Key("settings.theme.brand_color", "Brand colour")
	L10nThemeDangerColor          = i18n.Key("settings.theme.danger_color", "Danger colour")
	L10nThemeWarningColor         = i18n.Key("settings.theme.warning_color", "Warning colour")
	L10nThemeSuccessColor         = i18n.Key("settings.theme.success_color", "Success colour")
	L10nThemeInfoColor            = i18n.Key("settings.theme.info_color", "Information colour")
	L10nThemeFontSans             = i18n.Key("settings.theme.font_sans", "Text font")
	L10nThemeFontMono             = i18n.Key("settings.theme.font_mono", "Code font")
	L10nThemeRadius               = i18n.Key("settings.theme.radius", "Corner radius")
	L10nI18n                      = i18n.Key("settings.i18n", "Translations")
	L10nI18nDefaultLocale         = i18n.Key("settings.i18n.default_locale", "Locale for visitors without a chosen locale; empty means the source locale")
	L10nCFP                       = i18n.Key("settings.cfp", "Call for papers")
	L10nCFPClosed                 = i18n.Key("settings.cfp.closed", "Refuse new talks")
	L10nCFPMaxTalks               = i18n.Key("settings.cfp.max_talks", "Talks a speaker may give at one event")
	L10nCFPDefaultLevel           = i18n.Key("settings.cfp.default_level", "Level a new talk starts at")
	L10nCFPDefaultDuration        = i18n.Key("settings.cfp.default_duration", "Length a new talk starts at")
	L10nDemo                      = i18n.Key("settings.demo", "Demo data")
	L10nDemoEmail                 = i18n.Key("settings.demo.email", "Organiser's email address")
	L10nDemoName                  = i18n.Key("settings.demo.name", "Organiser's name")
	L10nDemoPassword              = i18n.Key("settings.demo.password", "Organiser's password")
)

// l10n holds every label that has a path.
var l10n = map[string]i18n.TKer{
	"log":                            L10nLog,
	"log.level":                      L10nLogLevel,
	"log.level.debug":                i18n.Key("settings.log.level.debug", "debug"),
	"log.level.info":                 i18n.Key("settings.log.level.info", "info"),
	"log.level.warn":                 i18n.Key("settings.log.level.warn", "warn"),
	"log.level.error":                i18n.Key("settings.log.level.error", "error"),
	"log.format":                     L10nLogFormat,
	"log.format.text":                i18n.Key("settings.log.format.text", "text"),
	"log.format.json":                i18n.Key("settings.log.format.json", "json"),
	"server":                         L10nServer,
	"server.addr":                    L10nServerAddr,
	"server.public_url":              L10nServerPublicURL,
	"server.shutdown_timeout":        L10nServerShutdownTimeout,
	"storage":                        L10nStorage,
	"storage.bolt":                   L10nStorageBolt,
	"storage.files":                  L10nStorageFiles,
	"db":                             L10nDB,
	"db.url":                         L10nDBURL,
	"mail":                           L10nMail,
	"mail.outbox":                    L10nMailOutbox,
	"auth":                           L10nAuth,
	"auth.issuer":                    L10nAuthIssuer,
	"auth.secret":                    L10nAuthSecret,
	"auth.secret_previous":           L10nAuthSecretPrevious,
	"auth.trusted_origins":           L10nAuthTrustedOrigins,
	"auth.session":                   L10nAuthSession,
	"auth.session.idle":              L10nAuthSessionIdle,
	"auth.session.remembered_idle":   L10nAuthSessionRememberedIdle,
	"auth.session.lifetime":          L10nAuthSessionLifetime,
	"auth.access":                    L10nAuthAccess,
	"auth.access.lifetime":           L10nAuthAccessLifetime,
	"auth.refresh":                   L10nAuthRefresh,
	"auth.refresh.grace":             L10nAuthRefreshGrace,
	"auth.flow":                      L10nAuthFlow,
	"auth.flow.lifetime":             L10nAuthFlowLifetime,
	"auth.fresh":                     L10nAuthFresh,
	"auth.fresh.window":              L10nAuthFreshWindow,
	"auth.code":                      L10nAuthCode,
	"auth.code.length":               L10nAuthCodeLength,
	"auth.code.lifetime":             L10nAuthCodeLifetime,
	"auth.code.tries":                L10nAuthCodeTries,
	"auth.code.resend":               L10nAuthCodeResend,
	"auth.code.hourly":               L10nAuthCodeHourly,
	"auth.link":                      L10nAuthLink,
	"auth.link.lifetime":             L10nAuthLinkLifetime,
	"auth.invitation":                L10nAuthInvitation,
	"auth.invitation.lifetime":       L10nAuthInvitationLifetime,
	"auth.device":                    L10nAuthDevice,
	"auth.device.lifetime":           L10nAuthDeviceLifetime,
	"auth.password":                  L10nAuthPassword,
	"auth.password.min_length":       L10nAuthPasswordMinLength,
	"auth.password.min_score":        L10nAuthPasswordMinScore,
	"auth.password.memory":           L10nAuthPasswordMemory,
	"auth.password.time":             L10nAuthPasswordTime,
	"auth.password.concurrency":      L10nAuthPasswordConcurrency,
	"auth.throttle":                  L10nAuthThrottle,
	"auth.throttle.free":             L10nAuthThrottleFree,
	"auth.throttle.ceiling":          L10nAuthThrottleCeiling,
	"auth.throttle.window":           L10nAuthThrottleWindow,
	"auth.cookie":                    L10nAuthCookie,
	"auth.cookie.insecure":           L10nAuthCookieInsecure,
	"theme":                          L10nTheme,
	"theme.brand_color":              L10nThemeBrandColor,
	"theme.danger_color":             L10nThemeDangerColor,
	"theme.warning_color":            L10nThemeWarningColor,
	"theme.success_color":            L10nThemeSuccessColor,
	"theme.info_color":               L10nThemeInfoColor,
	"theme.font_sans":                L10nThemeFontSans,
	"theme.font_mono":                L10nThemeFontMono,
	"theme.radius":                   L10nThemeRadius,
	"i18n":                           L10nI18n,
	"i18n.default_locale":            L10nI18nDefaultLocale,
	"cfp":                            L10nCFP,
	"cfp.closed":                     L10nCFPClosed,
	"cfp.max_talks":                  L10nCFPMaxTalks,
	"cfp.default_level":              L10nCFPDefaultLevel,
	"cfp.default_level.beginner":     i18n.Key("settings.cfp.default_level.beginner", "beginner"),
	"cfp.default_level.intermediate": i18n.Key("settings.cfp.default_level.intermediate", "intermediate"),
	"cfp.default_level.advanced":     i18n.Key("settings.cfp.default_level.advanced", "advanced"),
	"cfp.default_duration":           L10nCFPDefaultDuration,
	"demo":                           L10nDemo,
	"demo.email":                     L10nDemoEmail,
	"demo.name":                      L10nDemoName,
	"demo.password":                  L10nDemoPassword,
}

// L10n returns the label at path, or nil when there is none.
func L10n(path string) i18n.TKer { return l10n[path] }
