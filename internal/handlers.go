package internal

import (
	"net/http"
	"strconv"
	"time"
	"ws-whatever/ws"

	"github.com/labstack/echo"
	"gorm.io/gorm"
)

type CreateRoomRequest struct {
	Name        string `json:"name"`
	CommunityID int    `json:"community_id"`
	Type        string `json:"type"`
}

type RoomResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CommunityID int       `json:"community_id"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageResponse struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	ReplyToID *int      `json:"reply_to_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateRoom(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateRoomRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.Type != "group" && req.Type != "direct" {
			return echo.NewHTTPError(http.StatusBadRequest, "type must be 'group' or 'direct'")
		}

		if req.CommunityID == 0 {
			req.CommunityID = 1
		}

		room := ws.Room{
			Name:        req.Name,
			CommunityID: req.CommunityID,
			Type:        ws.RoomType(req.Type),
		}

		if err := db.Create(&room).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create room")
		}

		response := RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			CommunityID: room.CommunityID,
			Type:        string(room.Type),
			CreatedAt:   room.CreatedAt,
		}

		return c.JSON(http.StatusCreated, response)
	}
}

func ListRooms(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var rooms []ws.Room
		if err := db.Find(&rooms).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rooms")
		}

		response := make([]RoomResponse, len(rooms))
		for i, room := range rooms {
			response[i] = RoomResponse{
				ID:          room.ID,
				Name:        room.Name,
				CommunityID: room.CommunityID,
				Type:        string(room.Type),
				CreatedAt:   room.CreatedAt,
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetRoomMessages(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		roomID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
		}

		limit := 50
		if l := c.QueryParam("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		var messages []ws.Message
		err = db.
			Where("room_id = ? AND deleted_at IS NULL", roomID).
			Order("created_at DESC").
			Limit(limit).
			Find(&messages).Error
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch messages")
		}

		response := make([]MessageResponse, len(messages))
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			response[len(messages)-1-i] = MessageResponse{
				ID:        msg.ID,
				RoomID:    msg.RoomID,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				ReplyToID: msg.ReplyToID,
				CreatedAt: msg.CreatedAt,
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}

func DeleteMessage(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		messageID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid message id")
		}

		userID := c.Get("user_id")
		if userID == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var message ws.Message
		if err := db.First(&message, messageID).Error; err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}

		if message.SenderID != userID.(int) {
			return echo.NewHTTPError(http.StatusForbidden, "can only delete your own messages")
		}

		now := time.Now()
		message.DeletedAt = &now
		if err := db.Save(&message).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete message")
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func SearchMessages(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := c.QueryParam("q")
		if query == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
		}

		roomID := c.QueryParam("room_id")

		dbQuery := db.Where("content ILIKE ? AND deleted_at IS NULL", "%"+query+"%")

		if roomID != "" {
			if id, err := strconv.Atoi(roomID); err == nil {
				dbQuery = dbQuery.Where("room_id = ?", id)
			}
		}

		var messages []ws.Message
		if err := dbQuery.Order("created_at DESC").Limit(50).Find(&messages).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search messages")
		}

		response := make([]MessageResponse, len(messages))
		for i, msg := range messages {
			response[i] = MessageResponse{
				ID:        msg.ID,
				RoomID:    msg.RoomID,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				ReplyToID: msg.ReplyToID,
				CreatedAt: msg.CreatedAt,
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}
