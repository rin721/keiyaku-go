package i18n

import (
	"fmt"
	"sort"

	tooli18n "github.com/rin721/keiyaku-go/pkg/i18n"
	"github.com/rin721/keiyaku-go/types"
	"golang.org/x/text/language"
)

var translator = mustTranslator()

var enUSMessages = map[string]string{
	types.MessageOK:                      types.MessageOK,
	types.MessageInvalidArgument:         types.MessageInvalidArgument,
	types.MessageUnauthorized:            types.MessageUnauthorized,
	types.MessageForbidden:               types.MessageForbidden,
	types.MessageNotFound:                types.MessageNotFound,
	types.MessageConflict:                types.MessageConflict,
	types.MessageTooManyRequests:         types.MessageTooManyRequests,
	types.MessageInvalidCredential:       types.MessageInvalidCredential,
	types.MessageUserDisabled:            types.MessageUserDisabled,
	types.MessageInternal:                types.MessageInternal,
	types.MessageDependency:              types.MessageDependency,
	types.MessageRouteNotFound:           types.MessageRouteNotFound,
	types.MessageServiceUnavailable:      types.MessageServiceUnavailable,
	types.MessageMissingAuthClaims:       types.MessageMissingAuthClaims,
	types.MessageInvalidRequestBody:      types.MessageInvalidRequestBody,
	types.MessageInvalidAccessToken:      types.MessageInvalidAccessToken,
	types.MessageMissingAuthHeader:       types.MessageMissingAuthHeader,
	types.MessageInvalidAuthScheme:       types.MessageInvalidAuthScheme,
	types.MessagePermissionDenied:        types.MessagePermissionDenied,
	types.MessagePermissionNotReady:      types.MessagePermissionNotReady,
	types.MessagePermissionCheckFail:     types.MessagePermissionCheckFail,
	types.MessageAuthServiceNotReady:     types.MessageAuthServiceNotReady,
	types.MessageUserServiceNotReady:     types.MessageUserServiceNotReady,
	types.MessageArticleServiceNotReady:  types.MessageArticleServiceNotReady,
	types.MessageAuthHandlerNotReady:     types.MessageAuthHandlerNotReady,
	types.MessageUserHandlerNotReady:     types.MessageUserHandlerNotReady,
	types.MessageArticleHandlerNotReady:  types.MessageArticleHandlerNotReady,
	types.MessageInvalidUserID:           types.MessageInvalidUserID,
	types.MessageInvalidArticleID:        types.MessageInvalidArticleID,
	types.MessageMissingUser:             types.MessageMissingUser,
	types.MessagePasswordLength:          types.MessagePasswordLength,
	types.MessageUsernameExists:          types.MessageUsernameExists,
	types.MessageCheckUserFailed:         types.MessageCheckUserFailed,
	types.MessageLoadUserFailed:          types.MessageLoadUserFailed,
	types.MessageCreateUserFailed:        types.MessageCreateUserFailed,
	types.MessageAllocateUserIDFailed:    types.MessageAllocateUserIDFailed,
	types.MessageHashPasswordFailed:      types.MessageHashPasswordFailed,
	types.MessageVerifyPasswordFailed:    types.MessageVerifyPasswordFailed,
	types.MessageIssueTokenFailed:        types.MessageIssueTokenFailed,
	types.MessageAllocateArticleIDFailed: types.MessageAllocateArticleIDFailed,
	types.MessageCreateArticleFailed:     types.MessageCreateArticleFailed,
	types.MessageListArticlesFailed:      types.MessageListArticlesFailed,
}

var zhCNMessages = map[string]string{
	types.MessageOK:                      "成功",
	types.MessageInvalidArgument:         "参数错误",
	types.MessageUnauthorized:            "未认证",
	types.MessageForbidden:               "无权访问",
	types.MessageNotFound:                "资源不存在",
	types.MessageConflict:                "资源冲突",
	types.MessageTooManyRequests:         "请求过于频繁",
	types.MessageInvalidCredential:       "用户名或密码错误",
	types.MessageUserDisabled:            "用户已禁用",
	types.MessageInternal:                "服务器内部错误",
	types.MessageDependency:              "依赖服务错误",
	types.MessageRouteNotFound:           "路由不存在",
	types.MessageServiceUnavailable:      "服务暂时不可用",
	types.MessageMissingAuthClaims:       "缺少认证信息",
	types.MessageInvalidRequestBody:      "请求体无效",
	types.MessageInvalidAccessToken:      "访问令牌无效",
	types.MessageMissingAuthHeader:       "缺少授权请求头",
	types.MessageInvalidAuthScheme:       "授权类型无效",
	types.MessagePermissionDenied:        "权限不足",
	types.MessagePermissionNotReady:      "权限服务未就绪",
	types.MessagePermissionCheckFail:     "权限校验失败",
	types.MessageAuthServiceNotReady:     "认证服务未就绪",
	types.MessageUserServiceNotReady:     "用户服务未就绪",
	types.MessageArticleServiceNotReady:  "文章服务未就绪",
	types.MessageAuthHandlerNotReady:     "认证处理器未就绪",
	types.MessageUserHandlerNotReady:     "用户处理器未就绪",
	types.MessageArticleHandlerNotReady:  "文章处理器未就绪",
	types.MessageInvalidUserID:           "用户 ID 无效",
	types.MessageInvalidArticleID:        "文章 ID 无效",
	types.MessageMissingUser:             "缺少已认证用户",
	types.MessagePasswordLength:          "密码长度必须在 8 到 128 位之间",
	types.MessageUsernameExists:          "用户名已存在",
	types.MessageCheckUserFailed:         "检查用户失败",
	types.MessageLoadUserFailed:          "加载用户失败",
	types.MessageCreateUserFailed:        "创建用户失败",
	types.MessageAllocateUserIDFailed:    "分配用户 ID 失败",
	types.MessageHashPasswordFailed:      "生成密码哈希失败",
	types.MessageVerifyPasswordFailed:    "验证密码失败",
	types.MessageIssueTokenFailed:        "签发令牌失败",
	types.MessageAllocateArticleIDFailed: "分配文章 ID 失败",
	types.MessageCreateArticleFailed:     "创建文章失败",
	types.MessageListArticlesFailed:      "查询文章列表失败",
}

func mustTranslator() *tooli18n.Translator {
	translator, err := tooli18n.NewTranslator(tooli18n.Catalog{
		Default:   LanguageENUS,
		Supported: []language.Tag{LanguageENUS, LanguageZHCN},
		Messages: map[language.Tag][]tooli18n.Message{
			LanguageENUS: messagesFromMap(enUSMessages),
			LanguageZHCN: messagesFromMap(zhCNMessages),
		},
	})
	if err != nil {
		panic(fmt.Sprintf("build http i18n translator: %v", err))
	}
	return translator
}

func messagesFromMap(raw map[string]string) []tooli18n.Message {
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	messages := make([]tooli18n.Message, 0, len(keys))
	for _, key := range keys {
		messages = append(messages, tooli18n.Message{
			ID:    key,
			Other: raw[key],
		})
	}
	return messages
}
