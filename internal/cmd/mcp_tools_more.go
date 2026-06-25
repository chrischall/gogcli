package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpFormsTools returns the Google Forms MCP tool surface (service "forms").
func mcpFormsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpFormsGetTool(),
		mcpFormsListResponsesTool(),
		mcpFormsGetResponseTool(),
		mcpFormsCreateTool(),
		mcpFormsUpdateTool(),
		mcpFormsAddQuestionTool(),
	}
}

func mcpFormsGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_get",
		Service:     "forms",
		Risk:        mcpRiskRead,
		Description: "Get a Google Form definition.",
		Options: []mcp.ToolOption{
			mcp.WithString("form_id", mcp.Description("Form ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			formID, err := requireMCPString(req, "form_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "get").done(formID)
		},
	}
}

func mcpFormsListResponsesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_list_responses",
		Service:     "forms",
		Risk:        mcpRiskRead,
		Description: "List responses to a Google Form.",
		Options: []mcp.ToolOption{
			mcp.WithString("form_id", mcp.Description("Form ID"), mcp.Required()),
			mcp.WithString("filter", mcp.Description("Optional response filter")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			formID, err := requireMCPString(req, "form_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "responses", "list").
				str("filter", "--filter").
				num("max", "--max", 50, 1, 200).done(formID)
		},
	}
}

func mcpFormsGetResponseTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_get_response",
		Service:     "forms",
		Risk:        mcpRiskRead,
		Description: "Get a single Google Form response by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("form_id", mcp.Description("Form ID"), mcp.Required()),
			mcp.WithString("response_id", mcp.Description("Response ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			formID, err := requireMCPString(req, "form_id")
			if err != nil {
				return nil, err
			}
			responseID, err := requireMCPString(req, "response_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "responses", "get").done(formID, responseID)
		},
	}
}

func mcpFormsCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_create",
		Service:     "forms",
		Risk:        mcpRiskWrite,
		Description: "Create a Google Form. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Form title"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Form description")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "create", "--title", title).
				str("description", "--description").done()
		},
	}
}

func mcpFormsUpdateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_update",
		Service:     "forms",
		Risk:        mcpRiskWrite,
		Description: "Update a Google Form's title or description. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("form_id", mcp.Description("Form ID"), mcp.Required()),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("description", mcp.Description("New description")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			formID, err := requireMCPString(req, "form_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "update").
				str("title", "--title").
				str("description", "--description").done(formID)
		},
	}
}

func mcpFormsAddQuestionTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "forms_add_question",
		Service:     "forms",
		Risk:        mcpRiskWrite,
		Description: "Add a question to a Google Form. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("form_id", mcp.Description("Form ID"), mcp.Required()),
			mcp.WithString("title", mcp.Description("Question title"), mcp.Required()),
			mcp.WithString("type", mcp.Description("Question type"), mcp.Enum("text", "paragraph", "choice", "checkbox", "dropdown", "scale", "date", "time")),
			mcp.WithBoolean("required", mcp.Description("Mark the question required"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			formID, err := requireMCPString(req, "form_id")
			if err != nil {
				return nil, err
			}
			title, err := requireMCPText(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "forms", "questions", "add", "--title", title).
				str("type", "--type").
				flag("required", "--required").done(formID)
		},
	}
}

// mcpClassroomTools returns the Google Classroom MCP tool surface (service
// "classroom"). Classroom is exposed read-only.
func mcpClassroomTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpClassroomListCoursesTool(),
		mcpClassroomGetCourseTool(),
		mcpClassroomListCourseworkTool(),
		mcpClassroomGetCourseworkTool(),
		mcpClassroomListStudentsTool(),
		mcpClassroomListTeachersTool(),
		mcpClassroomRosterTool(),
		mcpClassroomListAnnouncementsTool(),
		mcpClassroomListSubmissionsTool(),
		mcpClassroomListTopicsTool(),
	}
}

func mcpClassroomListCoursesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_courses",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List Google Classroom courses.",
		Options: []mcp.ToolOption{
			mcp.WithString("state", mcp.Description("Course state filter"), mcp.Enum("ACTIVE", "ARCHIVED", "PROVISIONED", "DECLINED", "SUSPENDED")),
			mcp.WithString("teacher", mcp.Description("Filter by teacher (email or ID)")),
			mcp.WithString("student", mcp.Description("Filter by student (email or ID)")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "classroom", "courses", "list").
				str("state", "--state").
				str("teacher", "--teacher").
				str("student", "--student").
				num("max", "--max", 50, 1, 200).done()
		},
	}
}

func mcpClassroomGetCourseTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_get_course",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "Get a Google Classroom course by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "courses", "get").done(courseID)
		},
	}
}

func mcpClassroomListCourseworkTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_coursework",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List coursework for a Google Classroom course.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithString("state", mcp.Description("Coursework state filter")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "coursework", "list").
				str("state", "--state").
				num("max", "--max", 50, 1, 200).done(courseID)
		},
	}
}

func mcpClassroomGetCourseworkTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_get_coursework",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "Get a coursework item by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithString("coursework_id", mcp.Description("Coursework ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			courseworkID, err := requireMCPString(req, "coursework_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "coursework", "get").done(courseID, courseworkID)
		},
	}
}

func mcpClassroomListStudentsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_students",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List students enrolled in a course.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "students", "list").
				num("max", "--max", 50, 1, 200).done(courseID)
		},
	}
}

func mcpClassroomListTeachersTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_teachers",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List teachers for a course.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "teachers", "list").
				num("max", "--max", 50, 1, 200).done(courseID)
		},
	}
}

func mcpClassroomRosterTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_roster",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "Get a course roster (students and teachers).",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(100), mcp.Min(1), mcp.Max(400)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "roster").
				num("max", "--max", 100, 1, 400).done(courseID)
		},
	}
}

func mcpClassroomListAnnouncementsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_announcements",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List announcements for a course.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "announcements", "list").
				num("max", "--max", 50, 1, 200).done(courseID)
		},
	}
}

func mcpClassroomListSubmissionsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_submissions",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List student submissions for a coursework item.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithString("coursework_id", mcp.Description("Coursework ID"), mcp.Required()),
			mcp.WithString("state", mcp.Description("Submission state filter")),
			mcp.WithString("user", mcp.Description("Filter by student (email or ID)")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			courseworkID, err := requireMCPString(req, "coursework_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "submissions", "list").
				str("state", "--state").
				str("user", "--user").
				num("max", "--max", 50, 1, 200).done(courseID, courseworkID)
		},
	}
}

func mcpClassroomListTopicsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "classroom_list_topics",
		Service:     "classroom",
		Risk:        mcpRiskRead,
		Description: "List topics for a course.",
		Options: []mcp.ToolOption{
			mcp.WithString("course_id", mcp.Description("Course ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			courseID, err := requireMCPString(req, "course_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "classroom", "topics", "list").
				num("max", "--max", 50, 1, 200).done(courseID)
		},
	}
}

// mcpPhotosTools returns the Google Photos MCP tool surface (service "photos").
func mcpPhotosTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpPhotosListTool(),
		mcpPhotosGetTool(),
		mcpPhotosSearchTool(),
	}
}

func mcpPhotosListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "photos_list",
		Service:     "photos",
		Risk:        mcpRiskRead,
		Description: "List app-created Google Photos media items.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "photos", "list").
				num("max", "--max", 50, 1, 100).done()
		},
	}
}

func mcpPhotosGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "photos_get",
		Service:     "photos",
		Risk:        mcpRiskRead,
		Description: "Get an app-created Google Photos media item by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("media_item_id", mcp.Description("Media item ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			id, err := requireMCPString(req, "media_item_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "photos", "get").done(id)
		},
	}
}

func mcpPhotosSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "photos_search",
		Service:     "photos",
		Risk:        mcpRiskRead,
		Description: "Search app-created Google Photos media items.",
		Options: []mcp.ToolOption{
			mcp.WithString("album", mcp.Description("Album ID to search within")),
			mcp.WithString("from", mcp.Description("Start date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("End date (YYYY-MM-DD)")),
			mcp.WithString("media_type", mcp.Description("Media type filter"), mcp.Enum("ALL_MEDIA", "PHOTO", "VIDEO")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "photos", "search").
				str("album", "--album").
				str("from", "--from").
				str("to", "--to").
				str("media_type", "--media-type").
				num("max", "--max", 50, 1, 100).done()
		},
	}
}

// mcpMapsTools returns the Google Maps MCP tool surface (service "maps").
func mcpMapsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpMapsGeocodeTool(),
		mcpMapsReverseGeocodeTool(),
		mcpMapsDirectionsTool(),
		mcpMapsDistanceTool(),
		mcpMapsPlacesSearchTool(),
		mcpMapsPlaceDetailsTool(),
	}
}

func mcpMapsGeocodeTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_geocode",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Convert an address to coordinates.",
		Options: []mcp.ToolOption{
			mcp.WithString("address", mcp.Description("Address to geocode"), mcp.Required()),
			mcp.WithString("region", mcp.Description("Region bias (ccTLD)")),
			mcp.WithString("language", mcp.Description("Result language")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			address, err := requireMCPString(req, "address")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "geocode").
				str("region", "--region").
				str("language", "--language").done(address)
		},
	}
}

func mcpMapsReverseGeocodeTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_reverse_geocode",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Convert coordinates to an address.",
		Options: []mcp.ToolOption{
			mcp.WithString("lat", mcp.Description("Latitude"), mcp.Required()),
			mcp.WithString("lng", mcp.Description("Longitude"), mcp.Required()),
			mcp.WithString("language", mcp.Description("Result language")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			lat, err := requireMCPString(req, "lat")
			if err != nil {
				return nil, err
			}
			lng, err := requireMCPString(req, "lng")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "reverse-geocode", "--lat", lat, "--lng", lng).
				str("language", "--language").done()
		},
	}
}

func mcpMapsDirectionsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_directions",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Get directions between two locations.",
		Options: []mcp.ToolOption{
			mcp.WithString("origin", mcp.Description("Origin address or coordinates"), mcp.Required()),
			mcp.WithString("destination", mcp.Description("Destination address or coordinates"), mcp.Required()),
			mcp.WithString("mode", mcp.Description("Travel mode"), mcp.Enum("driving", "walking", "bicycling", "transit")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			origin, err := requireMCPString(req, "origin")
			if err != nil {
				return nil, err
			}
			destination, err := requireMCPString(req, "destination")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "directions", "--origin", origin, "--destination", destination).
				str("mode", "--mode").done()
		},
	}
}

func mcpMapsDistanceTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_distance",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Get a travel distance and duration matrix.",
		Options: []mcp.ToolOption{
			mcp.WithString("origins", mcp.Description("Origin(s), pipe-separated"), mcp.Required()),
			mcp.WithString("destinations", mcp.Description("Destination(s), pipe-separated"), mcp.Required()),
			mcp.WithString("mode", mcp.Description("Travel mode"), mcp.Enum("driving", "walking", "bicycling", "transit")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			origins, err := requireMCPString(req, "origins")
			if err != nil {
				return nil, err
			}
			destinations, err := requireMCPString(req, "destinations")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "distance", "--origins", origins, "--destinations", destinations).
				str("mode", "--mode").done()
		},
	}
}

func mcpMapsPlacesSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_places_search",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Search Google Maps Places by text.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search text"), mcp.Required()),
			mcp.WithString("region", mcp.Description("Region bias (ccTLD)")),
			mcp.WithString("language", mcp.Description("Result language")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "places", "search").
				str("region", "--region").
				str("language", "--language").done(query)
		},
	}
}

func mcpMapsPlaceDetailsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "maps_place_details",
		Service:     "maps",
		Risk:        mcpRiskRead,
		Description: "Get details for a Google Maps Place by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("place_id", mcp.Description("Place ID"), mcp.Required()),
			mcp.WithString("language", mcp.Description("Result language")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			placeID, err := requireMCPString(req, "place_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "maps", "places", "details").
				str("language", "--language").done(placeID)
		},
	}
}

// mcpYouTubeTools returns the YouTube Data API MCP tool surface (service "youtube").
func mcpYouTubeTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpYouTubeSearchTool(),
		mcpYouTubeListVideosTool(),
		mcpYouTubeListChannelsTool(),
		mcpYouTubeListPlaylistsTool(),
		mcpYouTubeListPlaylistItemsTool(),
		mcpYouTubeListCommentsTool(),
		mcpYouTubeCreatePlaylistTool(),
		mcpYouTubeAddToPlaylistTool(),
	}
}

func mcpYouTubeSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_search",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "Search YouTube for videos, channels, or playlists.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
			mcp.WithString("type", mcp.Description("Result type"), mcp.Enum("video", "channel", "playlist")),
			mcp.WithString("order", mcp.Description("Result order"), mcp.Enum("relevance", "date", "rating", "title", "viewCount")),
			mcp.WithString("channel_id", mcp.Description("Restrict to a channel ID")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "youtube", "search", "list").
				str("type", "--type").
				str("order", "--order").
				str("channel_id", "--channel-id").
				num("max", "--max", 25, 1, 50).done(query)
		},
	}
}

func mcpYouTubeListVideosTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_list_videos",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "List YouTube videos by ID or chart.",
		Options: []mcp.ToolOption{
			mcp.WithString("id", mcp.Description("Comma-separated video IDs")),
			mcp.WithString("chart", mcp.Description("Video chart"), mcp.Enum("mostPopular")),
			mcp.WithString("region", mcp.Description("Region code for charts")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "youtube", "videos", "list").
				str("id", "--id").
				str("chart", "--chart").
				str("region", "--region").
				num("max", "--max", 25, 1, 50).done()
		},
	}
}

func mcpYouTubeListChannelsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_list_channels",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "List YouTube channels by ID or for the authenticated user.",
		Options: []mcp.ToolOption{
			mcp.WithString("id", mcp.Description("Comma-separated channel IDs")),
			mcp.WithBoolean("mine", mcp.Description("List the authenticated user's channels"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "youtube", "channels", "list").
				str("id", "--id").
				flag("mine", "--mine").
				num("max", "--max", 25, 1, 50).done()
		},
	}
}

func mcpYouTubeListPlaylistsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_list_playlists",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "List YouTube playlists by channel or for the authenticated user.",
		Options: []mcp.ToolOption{
			mcp.WithString("channel_id", mcp.Description("Channel ID")),
			mcp.WithBoolean("mine", mcp.Description("List the authenticated user's playlists"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "youtube", "playlists", "list").
				str("channel_id", "--channel-id").
				flag("mine", "--mine").
				num("max", "--max", 25, 1, 50).done()
		},
	}
}

func mcpYouTubeListPlaylistItemsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_list_playlist_items",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "List the videos inside a YouTube playlist.",
		Options: []mcp.ToolOption{
			mcp.WithString("playlist_id", mcp.Description("Playlist ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			playlistID, err := requireMCPString(req, "playlist_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "youtube", "playlists", "items", "list", "--playlist-id", playlistID).
				num("max", "--max", 25, 1, 50).done()
		},
	}
}

func mcpYouTubeListCommentsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_list_comments",
		Service:     "youtube",
		Risk:        mcpRiskRead,
		Description: "List comment threads for a YouTube video or channel.",
		Options: []mcp.ToolOption{
			mcp.WithString("video_id", mcp.Description("Video ID")),
			mcp.WithString("channel_id", mcp.Description("Channel ID")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(50)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "youtube", "comments", "list").
				str("video_id", "--video-id").
				str("channel_id", "--channel-id").
				num("max", "--max", 25, 1, 50).done()
		},
	}
}

func mcpYouTubeCreatePlaylistTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_create_playlist",
		Service:     "youtube",
		Risk:        mcpRiskWrite,
		Description: "Create a YouTube playlist. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Playlist title"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Playlist description")),
			mcp.WithString("privacy", mcp.Description("Privacy status"), mcp.Enum("private", "unlisted", "public")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPText(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "youtube", "playlists", "create", "--title", title).
				str("description", "--description").
				str("privacy", "--privacy").done()
		},
	}
}

func mcpYouTubeAddToPlaylistTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "youtube_add_to_playlist",
		Service:     "youtube",
		Risk:        mcpRiskWrite,
		Description: "Add a video to a YouTube playlist. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("playlist_id", mcp.Description("Playlist ID"), mcp.Required()),
			mcp.WithString("video_id", mcp.Description("Video ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			playlistID, err := requireMCPString(req, "playlist_id")
			if err != nil {
				return nil, err
			}
			videoID, err := requireMCPString(req, "video_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "youtube", "playlists", "add", "--playlist-id", playlistID, "--video-id", videoID).done()
		},
	}
}

// mcpSearchConsoleTools returns the Search Console MCP tool surface (service
// "searchconsole").
func mcpSearchConsoleTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpSearchConsoleListSitesTool(),
		mcpSearchConsoleQueryTool(),
		mcpSearchConsoleListSitemapsTool(),
	}
}

func mcpSearchConsoleListSitesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "searchconsole_list_sites",
		Service:     "searchconsole",
		Risk:        mcpRiskRead,
		Description: "List accessible Search Console sites.",
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "searchconsole", "sites", "list").done()
		},
	}
}

//nolint:dupl // coincidental structural twin of calendar_search; unrelated domains, no shared helper warranted
func mcpSearchConsoleQueryTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "searchconsole_query",
		Service:     "searchconsole",
		Risk:        mcpRiskRead,
		Description: "Run a Search Analytics query for a site.",
		Options: []mcp.ToolOption{
			mcp.WithString("site_url", mcp.Description("Site URL (property)"), mcp.Required()),
			mcp.WithString("from", mcp.Description("Start date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("End date (YYYY-MM-DD)")),
			mcp.WithString("dimensions", mcp.Description("Comma-separated dimensions, e.g. query,page")),
			mcp.WithInteger("max", mcp.Description("Maximum rows"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(1000)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			siteURL, err := requireMCPString(req, "site_url")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "searchconsole", "query").
				str("from", "--from").
				str("to", "--to").
				str("dimensions", "--dimensions").
				num("max", "--max", 50, 1, 1000).done(siteURL)
		},
	}
}

func mcpSearchConsoleListSitemapsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "searchconsole_list_sitemaps",
		Service:     "searchconsole",
		Risk:        mcpRiskRead,
		Description: "List sitemaps for a Search Console site.",
		Options: []mcp.ToolOption{
			mcp.WithString("site_url", mcp.Description("Site URL (property)"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			siteURL, err := requireMCPString(req, "site_url")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "searchconsole", "sitemaps", "list").done(siteURL)
		},
	}
}
