package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"video-conference-be/internal/domain/room"
	"video-conference-be/pkg/utility"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

type RoomService interface {
	ListRooms(userID uint, username, role string) ([]room.Room, error)
	GetRoomByCode(code string) (*room.Room, error)
	CreateRoom(req room.Room) (*room.Room, error)
	UpdateRoom(req room.Room) (*room.Room, error)
	DeleteRoom(id uint) error
	BanUser(roomCode string, username string) error
	UnbanUser(roomCode string, username string) error
	UploadPresentation(ctx context.Context, roomID uint, file *multipart.FileHeader) (string, error)
	GetRoomByID(id uint) (*room.Room, error)
	DownloadPresentation(ctx context.Context, path string) (*minio.Object, string, error)
}

type roomService struct {
	minioClient *minio.Client
}

func NewRoomService() RoomService {
	c := utility.Config
	minioEndpoint := strings.TrimPrefix(c.MinioEndpoint, "http://")
	minioEndpoint = strings.TrimPrefix(minioEndpoint, "https://")

	mClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinioAccess, c.MinioSecret, ""),
		Secure: strings.HasPrefix(c.MinioEndpoint, "https"),
	})
	if err != nil {
		log.Printf("[WARNING] Failed to init MinIO client in RoomService: %v", err)
	}

	return &roomService{
		minioClient: mClient,
	}
}

func (s *roomService) ListRooms(userID uint, username, role string) ([]room.Room, error) {
	var rooms []room.Room

	query := utility.DB.
		Select("id", "name", "room_code", "group_id", "start_date", "end_date",
			"description", "max_participants", "assigned_to", "banned_users",
			"created_by_id", "created_at", "updated_at", "presentation_path").
		Preload("Group").
		Order("created_at ASC")

	if role != "admin" {
		query = query.Where(`
			(group_id IS NULL) OR 
			(group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)) OR
			(? = ANY(assigned_to))
		`, userID, username)
	}

	if err := query.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *roomService) GetRoomByCode(code string) (*room.Room, error) {
	var r room.Room
	if err := utility.DB.Preload("Group").
		Where("room_code = ?", code).
		First(&r).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &r, nil
}

func (s *roomService) GetRoomByID(id uint) (*room.Room, error) {
	var r room.Room
	if err := utility.DB.Preload("Group").First(&r, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &r, nil
}

func (s *roomService) CreateRoom(req room.Room) (*room.Room, error) {
	if req.Name == "" {
		return nil, errors.New("room name is required")
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return nil, errors.New("start date and end date are required")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	if req.GroupID != nil && *req.GroupID == 0 {
		req.GroupID = nil
	}

	req.RoomCode = utility.RandomToken(10)

	if err := utility.DB.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *roomService) UpdateRoom(req room.Room) (*room.Room, error) {
	if req.ID == 0 {
		return nil, errors.New("id is required")
	}

	var existingRoom room.Room
	if err := utility.DB.First(&existingRoom, req.ID).Error; err != nil {
		return nil, errors.New("room not found")
	}

	existingRoom.Name = req.Name
	existingRoom.Description = req.Description
	existingRoom.MaxParticipants = req.MaxParticipants
	existingRoom.AssignedTo = req.AssignedTo

	if req.GroupID != nil {
		if *req.GroupID == 0 {
			existingRoom.GroupID = nil
		} else {
			existingRoom.GroupID = req.GroupID
		}
	}

	if !req.StartDate.IsZero() {
		existingRoom.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		existingRoom.EndDate = req.EndDate
	}

	if existingRoom.EndDate.Before(existingRoom.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	if err := utility.DB.Save(&existingRoom).Error; err != nil {
		return nil, err
	}
	return &existingRoom, nil
}

func (s *roomService) DeleteRoom(id uint) error {
	if err := utility.DB.Delete(&room.Room{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (s *roomService) BanUser(roomCode string, username string) error {
	var r room.Room
	if err := utility.DB.Where("room_code = ?", roomCode).First(&r).Error; err != nil {
		return err
	}

	// Check if already banned
	for _, user := range r.BannedUsers {
		if user == username {
			return nil // Already banned
		}
	}

	r.BannedUsers = append(r.BannedUsers, username)
	return utility.DB.Save(&r).Error
}

func (s *roomService) UnbanUser(roomCode string, username string) error {
	var r room.Room
	if err := utility.DB.Where("room_code = ?", roomCode).First(&r).Error; err != nil {
		return err
	}

	newList := make([]string, 0, len(r.BannedUsers))
	for _, user := range r.BannedUsers {
		if user != username {
			newList = append(newList, user)
		}
	}

	r.BannedUsers = newList
	return utility.DB.Save(&r).Error
}

func (s *roomService) UploadPresentation(ctx context.Context, roomID uint, fileHeader *multipart.FileHeader) (string, error) {
	if s.minioClient == nil {
		return "", errors.New("minio client not initialized")
	}

	// 1. Get Room
	var r room.Room
	if err := utility.DB.First(&r, roomID).Error; err != nil {
		return "", errors.New("room not found")
	}

	// 2. Open File
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 3. Generate Path
	// presentations/{roomID}/{filename}
	ext := filepath.Ext(fileHeader.Filename)
	// simple sanitization
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, fileHeader.Filename)

	objectName := fmt.Sprintf("presentations/%d/%s", roomID, cleanName)
	contentType := "application/pdf"
	if ext != ".pdf" {
		// optional: force check or allow others
		// return "", errors.New("only pdf allowed")
    // for now let's allow it but prefer PDF
		contentType = fileHeader.Header.Get("Content-Type")
	}

	bucket := utility.Config.MinioBucket

	// 4. Upload
	_, err = s.minioClient.PutObject(ctx, bucket, objectName, src, fileHeader.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %v", err)
	}

	// 5. Construct URL
	// similar logic to recording service
	base := strings.TrimRight(utility.Config.MinioEndpoint, "/")
	if !strings.HasPrefix(base, "http") {
		// if config doesn't have scheme, prepend https (or http based on secure)
		// Assuming config usually has it, if not default to https
		base = "https://" + base
	}
	// Note: Config.MinioEndpoint might be internal k8s dns.
	// For public access, we might need a different public URL if configured.
	// But let's stick to what RecordingService uses or just store the relative path if client constructs it?
	// RecordingService stores full link: link := fmt.Sprintf("%s/%s/%s", base, bucket, filePath)
    // We will do the same.

    // Update: user Config.MinioBaseURL if it exists (it was in config.go)
    urlBase := utility.Config.MinioBaseURL
    if urlBase == "" {
        urlBase = utility.Config.MinioEndpoint
    }
    urlBase = strings.TrimRight(urlBase, "/")
    bucketName := strings.Trim(utility.Config.MinioBucket, "/")
    
	fullURL := fmt.Sprintf("%s/%s/%s", urlBase, bucketName, objectName)

	// 6. Update DB
	r.PresentationPath = fullURL
	if err := utility.DB.Save(&r).Error; err != nil {
		return "", err
	}

	return fullURL, nil
}

func (s *roomService) DownloadPresentation(ctx context.Context, path string) (*minio.Object, string, error) {
	bucket := utility.Config.MinioBucket
	objectName := path

	// If path is a full URL, try to extract the object name
	// Format: .../bucketName/objectName
	// Or .../bucketName/presentations/id/file
	if strings.HasPrefix(path, "http") {
		// Try to find the bucket name and split after it
		// This is a bit heuristic.
		// Assuming standard format: base/bucket/key
		parts := strings.Split(path, bucket+"/")
		if len(parts) > 1 {
			objectName = parts[1]
		} else {
			// Fallback: assume the last parts are the key
			// presentations/id/filename
			if idx := strings.Index(path, "presentations/"); idx != -1 {
				objectName = path[idx:]
			}
		}
	}

	obj, err := s.minioClient.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	// Verify object exists by stating it
	stat, err := obj.Stat()
	if err != nil {
		return nil, "", err // likely 404
	}

	return obj, stat.ContentType, nil
}
