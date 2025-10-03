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

type AddParticipantRequest struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

func AddRoomParticipant(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		roomID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
		}

		var req AddParticipantRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.UserID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
		}

		if req.Role == "" {
			req.Role = "member"
		}

		var room ws.Room
		if err := db.First(&room, roomID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "room not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch room")
		}

		var user ws.User
		if err := db.First(&user, req.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				user = ws.User{ID: req.UserID}
				if err := db.Create(&user).Error; err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
				}
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user")
			}
		}

		participant := ws.RoomParticipant{
			RoomID: roomID,
			UserID: req.UserID,
			Role:   req.Role,
		}

		if err := db.Create(&participant).Error; err != nil {
			return echo.NewHTTPError(http.StatusConflict, "user already in room")
		}

		return c.JSON(http.StatusCreated, map[string]string{"status": "user added to room"})
	}
}

type CreateDirectMessageRequest struct {
	CommunityID int `json:"community_id"`
	UserID      int `json:"user_id"`
}

type DirectMessageResponse struct {
	RoomID      int       `json:"room_id"`
	CommunityID int       `json:"community_id"`
	UserAID     int       `json:"user_a_id"`
	UserBID     int       `json:"user_b_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func CreateOrGetDirectMessage(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id")
		if userID == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var req CreateDirectMessageRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.UserID == 0 || req.CommunityID == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "user_id and community_id are required")
		}

		currentUserID := userID.(int)
		if currentUserID == req.UserID {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot create DM with yourself")
		}

		// Check if DM room already exists by finding a direct room with these 2 participants
		var existingRoom ws.Room
		err := db.Where("community_id = ? AND type = ?", req.CommunityID, ws.RoomTypeDirect).
		 Joins("JOIN room_participants rp1 ON rp1.room_id = rooms.id AND rp1.user_id = ?", currentUserID).
		Joins("JOIN room_participants rp2 ON rp2.room_id = rooms.id AND rp2.user_id = ?", req.UserID).
		 Where("(SELECT COUNT(*) FROM room_participants WHERE room_id = rooms.id) = 2").
		 First(&existingRoom).Error

	if err == nil {
		 return c.JSON(http.StatusOK, DirectMessageResponse{
		 RoomID:      existingRoom.ID,
		 CommunityID: existingRoom.CommunityID,
		 UserAID:     min(currentUserID, req.UserID),
		UserBID:     max(currentUserID, req.UserID),
		CreatedAt:   existingRoom.CreatedAt,
		})
		}

		if err != gorm.ErrRecordNotFound {
		 return echo.NewHTTPError(http.StatusInternalServerError, "failed to check existing DM")
	}

		tx := db.Begin()
		defer func() {
		if r := recover(); r != nil {
		tx.Rollback()
		}
		}()

		for _, uid := range []int{currentUserID, req.UserID} {
		var user ws.User
		if err := tx.First(&user, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
		user = ws.User{ID: uid}
		if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
		} else {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user")
		}
		}
		}

		room := ws.Room{
		Name:        "",
		CommunityID: req.CommunityID,
		Type:        ws.RoomTypeDirect,
		}
		if err := tx.Create(&room).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create room")
		}

		participantA := ws.RoomParticipant{
		RoomID: room.ID,
		UserID: currentUserID,
		Role:   "member",
		}
		participantB := ws.RoomParticipant{
		 RoomID: room.ID,
		UserID: req.UserID,
		Role:   "member",
		}

		if err := tx.Create(&participantA).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add participants")
		}
		if err := tx.Create(&participantB).Error; err != nil {
		 tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add participants")
		}

		if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
		}

		return c.JSON(http.StatusCreated, DirectMessageResponse{
		RoomID:      room.ID,
		 CommunityID: req.CommunityID,
		UserAID:     min(currentUserID, req.UserID),
		UserBID:     max(currentUserID, req.UserID),
		 CreatedAt:   room.CreatedAt,
		})
	}
}

func GetUserRooms(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id")
		if userID == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var participants []ws.RoomParticipant
		if err := db.Where("user_id = ?", userID.(int)).
			Preload("Room").
			Find(&participants).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rooms")
		}

		response := make([]RoomResponse, len(participants))
		for i, p := range participants {
			response[i] = RoomResponse{
				ID:          p.Room.ID,
				Name:        p.Room.Name,
				CommunityID: p.Room.CommunityID,
				Type:        string(p.Room.Type),
				CreatedAt:   p.Room.CreatedAt,
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}
