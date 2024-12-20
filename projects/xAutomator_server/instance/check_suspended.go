package instance

import (
	"fmt"
	"io"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

// CheckSuspended check if account is suspended
// returns success status, error type, logs, username and creation date
func (twitter *Twitter) CheckSuspended() (bool, string, []string, string, string) {
	var logs []string
	errorType := "Unknown"
	var creationDate string
	checkSuspendedURL := "https://api.x.com/1.1/live_pipeline/update_subscriptions"

	var stringData = "sub_topics=&unsub_topics=&"
	data := strings.NewReader(stringData)

	for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
		// Create new request
		req, err := http.NewRequest("POST", checkSuspendedURL, data)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to build check suspended request: %s", err.Error()))
			continue
		}
		req.Header = http.Header{
			"accept":                {"*/*"},
			"accept-encoding":       {"gzip, deflate, br"},
			"authorization":         {twitter.queryID.BearerToken},
			"content-type":          {"application/x-www-form-urlencoded"},
			"cookie":                {twitter.cookies.CookiesToHeader()},
			"origin":                {"https://twitter.com"},
			"sec-ch-ua-mobile":      {"?0"},
			"sec-ch-ua-platform":    {`"Windows"`},
			"sec-fetch-dest":        {"empty"},
			"sec-fetch-mode":        {"cors"},
			"sec-fetch-site":        {"same-origin"},
			"x-csrf-token":          {twitter.ct0},
			"x-twitter-active-user": {"yes"},
			"x-twitter-auth-type":   {"OAuth2Session"},
			http.HeaderOrderKey: {
				"accept",
				"accept-encoding",
				"authorization",
				"content-type",
				"cookie",
				"origin",
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
			logs = append(logs, fmt.Sprintf("Failed to do check suspended request: %s", err.Error()))
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to read check suspended response body: %s", err.Error()))
			continue
		}

		bodyString := string(bodyBytes)

		// if resp.StatusCode >= 200 && resp.StatusCode <= 299 && strings.Contains(bodyString, "screen_name") {
		// 	var responseValid GetValidUserResponse
		// 	err = json.Unmarshal(bodyBytes, &responseValid)
		// 	if err != nil {
		// 		logs = append(logs, fmt.Sprintf("Failed to do unmarshal in check valid response: %s", err.Error()))
		// 		continue
		// 	}
		// 	creationDate = responseValid[0].CreatedAt
		// 	layout := "Mon Jan 02 15:04:05 -0700 2006"
		// 	creationDateStr, err := time.Parse(layout, creationDate)
		// 	creationDate = creationDateStr.Format("01/2006")
		// 	if err != nil {
		// 		logs = append(logs, fmt.Sprintf("Error parsing account creation date: %v", err))
		// 		return false, errorType, logs, "", creationDate
		// 	}

		// 	if responseValid[0].Suspended {
		// 		logs = append(logs, fmt.Sprintf("%s is suspended. Created at %s", twitter.Username, creationDate))
		// 		return false, "Suspended", logs, "", creationDate
		// 	} else {
		// 		logs = append(logs, fmt.Sprintf("%s is not suspended. Created at %s", twitter.Username, creationDate))
		// 		return true, "", logs, twitter.Username, creationDate
		// 	}

		// } else
		if strings.Contains(bodyString, "this account is temporarily locked") {
			logs = append(logs, "Account is temporarily locked")
			return false, "Locked", logs, "", creationDate

		} else if strings.Contains(bodyString, "Could not authenticate you") {
			logs = append(logs, "Could not authenticate you")
			return false, "Unauthenticated", logs, "", creationDate
		} else if strings.Contains(bodyString, "Your account is suspended") {
			logs = append(logs, "Account is suspended")
			return false, "Suspended", logs, "", creationDate
		} else {
			logs = append(logs, fmt.Sprintf("Unknown response while check suspended: %s", bodyString))
			continue
		}
	}

	logs = append(logs, fmt.Sprintf("Unable to check if %s is suspended", twitter.Username))
	return false, errorType, logs, "", creationDate
}

type GetValidUserResponse []struct {
	ID          int64  `json:"id"`
	IDStr       string `json:"id_str"`
	Name        string `json:"name"`
	ScreenName  string `json:"screen_name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	URL         any    `json:"url"`
	Entities    struct {
		Description struct {
			Urls []any `json:"urls"`
		} `json:"description"`
	} `json:"entities"`
	Protected            bool   `json:"protected"`
	FollowersCount       int    `json:"followers_count"`
	FastFollowersCount   int    `json:"fast_followers_count"`
	NormalFollowersCount int    `json:"normal_followers_count"`
	FriendsCount         int    `json:"friends_count"`
	ListedCount          int    `json:"listed_count"`
	CreatedAt            string `json:"created_at"`
	FavouritesCount      int    `json:"favourites_count"`
	UtcOffset            any    `json:"utc_offset"`
	TimeZone             any    `json:"time_zone"`
	GeoEnabled           bool   `json:"geo_enabled"`
	Verified             bool   `json:"verified"`
	StatusesCount        int    `json:"statuses_count"`
	MediaCount           int    `json:"media_count"`
	Lang                 any    `json:"lang"`
	Status               struct {
		CreatedAt string `json:"created_at"`
		ID        int64  `json:"id"`
		IDStr     string `json:"id_str"`
		Text      string `json:"text"`
		Truncated bool   `json:"truncated"`
		Entities  struct {
			Hashtags     []any `json:"hashtags"`
			Symbols      []any `json:"symbols"`
			UserMentions []any `json:"user_mentions"`
			Urls         []any `json:"urls"`
		} `json:"entities"`
		Source               string `json:"source"`
		InReplyToStatusID    any    `json:"in_reply_to_status_id"`
		InReplyToStatusIDStr any    `json:"in_reply_to_status_id_str"`
		InReplyToUserID      any    `json:"in_reply_to_user_id"`
		InReplyToUserIDStr   any    `json:"in_reply_to_user_id_str"`
		InReplyToScreenName  any    `json:"in_reply_to_screen_name"`
		Geo                  any    `json:"geo"`
		Coordinates          any    `json:"coordinates"`
		Place                any    `json:"place"`
		Contributors         any    `json:"contributors"`
		IsQuoteStatus        bool   `json:"is_quote_status"`
		RetweetCount         int    `json:"retweet_count"`
		FavoriteCount        int    `json:"favorite_count"`
		Favorited            bool   `json:"favorited"`
		Retweeted            bool   `json:"retweeted"`
		Lang                 string `json:"lang"`
		SupplementalLanguage any    `json:"supplemental_language"`
	} `json:"status"`
	ContributorsEnabled            bool   `json:"contributors_enabled"`
	IsTranslator                   bool   `json:"is_translator"`
	IsTranslationEnabled           bool   `json:"is_translation_enabled"`
	ProfileBackgroundColor         string `json:"profile_background_color"`
	ProfileBackgroundImageURL      any    `json:"profile_background_image_url"`
	ProfileBackgroundImageURLHTTPS any    `json:"profile_background_image_url_https"`
	ProfileBackgroundTile          bool   `json:"profile_background_tile"`
	ProfileImageURL                string `json:"profile_image_url"`
	ProfileImageURLHTTPS           string `json:"profile_image_url_https"`
	ProfileLinkColor               string `json:"profile_link_color"`
	ProfileSidebarBorderColor      string `json:"profile_sidebar_border_color"`
	ProfileSidebarFillColor        string `json:"profile_sidebar_fill_color"`
	ProfileTextColor               string `json:"profile_text_color"`
	ProfileUseBackgroundImage      bool   `json:"profile_use_background_image"`
	HasExtendedProfile             bool   `json:"has_extended_profile"`
	DefaultProfile                 bool   `json:"default_profile"`
	DefaultProfileImage            bool   `json:"default_profile_image"`
	PinnedTweetIds                 []any  `json:"pinned_tweet_ids"`
	PinnedTweetIdsStr              []any  `json:"pinned_tweet_ids_str"`
	HasCustomTimelines             bool   `json:"has_custom_timelines"`
	CanMediaTag                    bool   `json:"can_media_tag"`
	FollowedBy                     bool   `json:"followed_by"`
	Following                      bool   `json:"following"`
	FollowRequestSent              bool   `json:"follow_request_sent"`
	Notifications                  bool   `json:"notifications"`
	AdvertiserAccountType          string `json:"advertiser_account_type"`
	AdvertiserAccountServiceLevels []any  `json:"advertiser_account_service_levels"`
	BusinessProfileState           string `json:"business_profile_state"`
	TranslatorType                 string `json:"translator_type"`
	WithheldInCountries            []any  `json:"withheld_in_countries"`
	Suspended                      bool   `json:"suspended"`
	NeedsPhoneVerification         bool   `json:"needs_phone_verification"`
	RequireSomeConsent             bool   `json:"require_some_consent"`
}
