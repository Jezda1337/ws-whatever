package ws

import "time"

type RoomType string

const (
	RoomTypeGroup  RoomType = "group"
	RoomTypeDirect RoomType = "direct"
)

type Message struct {
	ID        int       `gorm:"primaryKey"`
	RoomID    int       `gorm:"not null;index:idx_messages_room_created_at"`
	SenderID  int       `gorm:"not null"`
	Content   string    `gorm:"type:text;not null"`
	ReplyToID *int      `gorm:"index:idx_messages_reply_to_id"`
	IsPinned  bool      `gorm:"default:false;index:idx_messages_room_pinned"`
	IsEdited  bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"default:now();index:idx_messages_room_created_at"`
	UpdatedAt *time.Time
	DeletedAt *time.Time
	Room      Room     `gorm:"foreignKey:RoomID"`
	Sender    User     `gorm:"foreignKey:SenderID"`
	ReplyTo   *Message `gorm:"foreignKey:ReplyToID"`
}

type MessageAttachment struct {
	ID        int    `gorm:"primaryKey"`
	MessageID int    `gorm:"not null"`
	FilePath  string `gorm:"type:text;not null"`
	FileType  string `gorm:"type:text;not null"`
	FileSize  int
	FileMime  string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:now()"`
	Message   Message   `gorm:"foreignKey:MessageID"`
}

type Room struct {
	ID          int       `gorm:"primaryKey"`
	Name        string    `gorm:"type:text"`
	CommunityID int       `gorm:"not null;index:idx_rooms_event"`
	Type        RoomType  `gorm:"type:room_type;not null"`
	CreatedAt   time.Time `gorm:"default:now()"`
	Community   Community `gorm:"foreignKey:CommunityID"`
}

type RoomParticipant struct {
	ID       int       `gorm:"primaryKey"`
	RoomID   int       `gorm:"not null;uniqueIndex:idx_room_participants_room_user;index:idx_room_participants_room"`
	UserID   int       `gorm:"not null;uniqueIndex:idx_room_participants_room_user;index:idx_room_participants_user"`
	Role     string    `gorm:"type:varchar(50);not null"`
	JoinedAt time.Time `gorm:"default:now()"`
	Room     Room      `gorm:"foreignKey:RoomID"`
	User     User      `gorm:"foreignKey:UserID"`
}

type MessageRead struct {
	ID        int `gorm:"primaryKey"`
	MessageID int `gorm:"not null;uniqueIndex:idx_message_reads_message_user"`
	UserID    int `gorm:"not null;uniqueIndex:idx_message_reads_message_user;index:idx_message_reads_user"`
	ReadAt    *time.Time
	Message   Message `gorm:"foreignKey:MessageID"`
	User      User    `gorm:"foreignKey:UserID"`
}

type MessageReaction struct {
	ID           int       `gorm:"primaryKey"`
	MessageID    int       `gorm:"not null;index:idx_message_reactions_message;uniqueIndex:uniq_message_reactions_user_type"`
	UserID       int       `gorm:"not null;uniqueIndex:idx_message_reactions_message_user;uniqueIndex:uniq_message_reactions_user_type"`
	ReactionType string    `gorm:"type:varchar(50);not null;uniqueIndex:uniq_message_reactions_user_type"`
	CreatedAt    time.Time `gorm:"default:now()"`
	Message      Message   `gorm:"foreignKey:MessageID"`
	User         User      `gorm:"foreignKey:UserID"`
}

type Community struct {
	ID int `gorm:"primaryKey"`
}

type User struct {
	ID int `gorm:"primaryKey"`
}
