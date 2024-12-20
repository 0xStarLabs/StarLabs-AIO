package instance

import (
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"

	"io"
	"strings"
	"twitter/extra"
	"twitter/instance/additional_twitter_methods"
)

// Comment perform comment action
func (twitter *Twitter) Comment(tweetContent string, tweetLink string) (bool, string, []string) {
    var logs []string
    errorType := "Unknown"

    tweetURL := fmt.Sprintf("https://twitter.com/i/api/graphql/%s/CreateTweet", twitter.queryID.TweetID)
    tweetID := additional_twitter_methods.GetTweetID(tweetLink)
    if tweetID == "" {
        logs = append(logs, "Invalid tweet link")
        return false, errorType, logs
    }

    for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
        var stringData = fmt.Sprintf(`{"variables":{"tweet_text":"%s","reply":{"in_reply_to_tweet_id":"%s","exclude_reply_user_ids":[]},"dark_request":false,"media":{"media_entities":[],"possibly_sensitive":false},"semantic_annotation_ids":[],"disallowed_reply_options":null},"features":{"tweetypie_unmention_optimization_enabled":true,"responsive_web_media_download_video_enabled":true,"communities_web_enable_tweet_community_results_fetch":true,"c9s_tweet_anatomy_moderator_badge_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":true,"tweet_awards_web_tipping_enabled":false,"creator_subscriptions_quote_tweet_preview_enabled":false,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"articles_preview_enabled":true,"rweb_video_timestamps_enabled":true,"rweb_tipjar_consumption_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false},"queryId":"%s"}`, tweetContent, tweetID, twitter.queryID.TweetID)
        data := strings.NewReader(stringData)

        req, err := http.NewRequest("POST", tweetURL, data)
        if err != nil {
            logs = append(logs, fmt.Sprintf("Failed to build comment request: %s", err.Error()))
            continue
        }

        // Set headers (same as in original)
        req.Header = http.Header{
            "accept":                {"*/*"},
            "accept-encoding":       {"gzip, deflate, br"},
            "authorization":         {twitter.queryID.BearerToken},
            "content-type":          {"application/json"},
            "cookie":                {twitter.cookies.CookiesToHeader()},
            "origin":                {"https://twitter.com"},
            "referer":               {"https://twitter.com/compose/tweet"},
            "sec-ch-ua-mobile":      {"?0"},
            "sec-ch-ua-platform":    {`"Windows"`},
            "sec-fetch-dest":        {"empty"},
            "sec-fetch-mode":        {"cors"},
            "sec-fetch-site":        {"same-origin"},
            "x-csrf-token":          {twitter.ct0},
            "x-twitter-active-user": {"yes"},
            "x-twitter-auth-type":   {"OAuth2Session"},
            http.PHeaderOrderKey:    {":authority", ":method", ":path", ":scheme"},
        }

        resp, err := twitter.client.Do(req)
        if err != nil {
            logs = append(logs, fmt.Sprintf("Failed to do comment request: %s", err.Error()))
            continue
        }
        defer resp.Body.Close()

        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            logs = append(logs, fmt.Sprintf("Failed to read comment response body: %s", err.Error()))
            continue
        }

        bodyString := string(bodyBytes)

        if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
            if strings.Contains(bodyString, "duplicate") {
                logs = append(logs, fmt.Sprintf("%s already commented this tweet", twitter.Username))
                return true, "", logs
            } else if strings.Contains(bodyString, "rest_id") {
                logs = append(logs, fmt.Sprintf("%s successfully commented", twitter.Username))
                return true, "", logs
            }
        } else if strings.Contains(bodyString, "this account is temporarily locked") {
            logs = append(logs, "Account is temporarily locked")
            return false, "Locked", logs
        } else if strings.Contains(bodyString, "Could not authenticate you") {
            logs = append(logs, "Could not authenticate you")
            return false, "Unauthenticated", logs
        } else {
            logs = append(logs, fmt.Sprintf("Unknown response while comment: %s", bodyString))
            continue
        }
    }

    logs = append(logs, "Unable to do comment")
    return false, errorType, logs
}

func (twitter *Twitter) CommentWithPicture(tweetContent string, tweetLink string, pictureBase64Encoded string) bool {
	tweetURL := fmt.Sprintf("https://twitter.com/i/api/graphql/%s/CreateTweet", twitter.queryID.TweetID)
	tweetID := additional_twitter_methods.GetTweetID(tweetLink)
	if tweetID == "" {
		return false
	}

	for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
		mediaID := twitter.UploadImageToTwitter(pictureBase64Encoded)
		if mediaID == "" {
			return false
		}

		var stringData = fmt.Sprintf(`{"variables":{"tweet_text":"%s","reply":{"in_reply_to_tweet_id":"%s","exclude_reply_user_ids":[]},"dark_request":false,"media":{"media_entities":[{"media_id":"%s","tagged_users":[]}],"possibly_sensitive":false},"semantic_annotation_ids":[]},"features":{"c9s_tweet_anatomy_moderator_badge_enabled":true,"tweetypie_unmention_optimization_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":false,"tweet_awards_web_tipping_enabled":false,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"rweb_video_timestamps_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"responsive_web_media_download_video_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false},"queryId":"%s"}`, tweetContent, tweetID, mediaID, twitter.queryID.TweetID)
		data := strings.NewReader(stringData)

		// Create new request
		req, err := http.NewRequest("POST", tweetURL, data)
		if err != nil {
			extra.Logger{}.Error("Failed to build comment with picture request: %s", err.Error())
			continue
		}
		req.Header = http.Header{
			"accept":                {"*/*"},
			"accept-encoding":       {"gzip, deflate, br"},
			"authorization":         {twitter.queryID.BearerToken},
			"content-type":          {"application/json"},
			"cookie":                {twitter.cookies.CookiesToHeader()},
			"origin":                {"https://twitter.com"},
			"referer":               {"https://twitter.com/compose/tweet"},
			"sec-ch-ua-mobile":      {"?0"},
			"sec-ch-ua-platform":    {`"Windows"`},
			"sec-fetch-dest":        {"empty"},
			"sec-fetch-mode":        {"cors"},
			"sec-fetch-site":        {"same-origin"},
			"x-csrf-token":          {twitter.ct0},
			"x-twitter-active-user": {"yes"},
			"x-twitter-auth-type":   {"OAuth2Session"},
			// x-client-transaction-id
			http.HeaderOrderKey: {
				"accept",
				"accept-encoding",
				"authorization",
				"content-type",
				"cookie",
				"origin",
				"referer",
				"sec-ch-ua-mobile",
				"sec-ch-ua-platform",
				"sec-fetch-dest",
				"sec-fetch-mode",
				"sec-fetch-site",
				"user-agent",
				"x-csrf-token",
				"x-twitter-active-user",
				"x-twitter-auth-type",
			},
			http.PHeaderOrderKey: {":authority", ":method", ":path", ":scheme"},
		}

		resp, err := twitter.client.Do(req)
		if err != nil {
			extra.Logger{}.Error("Failed to do comment with picture request: %s", err.Error())
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			extra.Logger{}.Error("Failed to read comment with picture response body: %s", err.Error())
			continue
		}

		bodyString := string(bodyBytes)

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			if strings.Contains(bodyString, "rest_id") {
				var responseTweet tweetResponse
				err = json.Unmarshal(bodyBytes, &responseTweet)
				if err != nil {
					extra.Logger{}.Error("%d | Failed to do unmarshal in comment with picture response: %s", twitter.index, err.Error())
					continue
				}
				extra.Logger{}.Success("%d | %s successfully commented with picture", twitter.index, twitter.Username)
				return true
			}

		} else if strings.Contains(bodyString, "this account is temporarily locked") {
			extra.Logger{}.Error("%d | Account is temporarily locked", twitter.index)
			return false

		} else if strings.Contains(bodyString, "Could not authenticate you") {
			extra.Logger{}.Error("%d | Could not authenticate you", twitter.index)
			return false
		} else {
			extra.Logger{}.Error("%d | Unknown response while comment with picture: %s", twitter.index, bodyString)
			continue
		}
	}

	extra.Logger{}.Error("%d | Unable to do comment with picture", twitter.index)
	return false
}
