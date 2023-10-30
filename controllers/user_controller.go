package controllers

import (
	"net/http"
	"time"

	"gihub.com/amit/Go-Gist/models"
	"gihub.com/amit/Go-Gist/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(DB *gorm.DB) UserController {
	return UserController{
		DB: DB,
	}
}

func (uc *UserController) GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	userResponse := &models.UserResponse{
		Username:     currentUser.Username,
		FirstName:    currentUser.FirstName,
		LastName:     currentUser.LastName,
		Email:        currentUser.Email,
		Role:         currentUser.Role,
		Provider:     currentUser.Provider,
		UserMetadata: currentUser.UserMetadata,
		CreatedAt:    currentUser.CreatedAt,
		UpdatedAt:    currentUser.UpdatedAt,
		Gists:        currentUser.Gists,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

func (uc *UserController) GetUser(ctx *gin.Context) {
	username := ctx.Params.ByName("username")

	var user models.User
	result := uc.DB.Preload("UserMetadata").First(&user, "username = ?", username)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "user with username: '" + username + "' does not exist"})
		return
	}

	gists := make([]models.Gist, 0)
	for _, gist := range user.Gists {
		if !gist.Private {
			gists = append(gists, gist)
		}
	}

	publicUserProfile := models.PublicUserProfileResponse{
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		UserMetadata: user.UserMetadata,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": publicUserProfile}})
}

func (uc *UserController) GetUserGists(ctx *gin.Context) {
	username := ctx.Params.ByName("username")

	var user models.User
	result := uc.DB.
		Preload("Gists").
		Preload("Gists.GistContent").
		First(&user, "username = ?", username)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "user with username: '" + username + "' does not exist"})
		return
	}

	gists := make([]models.Gist, 0)
	for _, gist := range user.Gists {
		if !gist.Private {
			gists = append(gists, gist)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"username": username, "gists": gists}})
}

func (uc *UserController) GetUserGistsIds(ctx *gin.Context) {
	username := ctx.Params.ByName("username")

	var user models.User
	result := uc.DB.
		Preload("Gists").
		First(&user, "username = ?", username)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "user with username: '" + username + "' does not exist"})
		return
	}

	gistIds := make([]uuid.UUID, 0)
	for _, gist := range user.Gists {
		if !gist.Private {
			gistIds = append(gistIds, gist.ID)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"username": username, "gistIds": gistIds}})
}

func (uc *UserController) CreateGist(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.CreateGistRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	now := time.Now()

	gists := currentUser.Gists
	for _, gist := range gists {
		if gist.Name == payload.Name {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Gist with name: '" + payload.Name + "' already exists"})
			return
		}
	}

	newGist := models.Gist{
		Username:    currentUser.Username,
		Private:     payload.Private,
		GistContent: models.GistContent{
			Content: payload.Content,
		},
		Name:        payload.Name,
		Title:       payload.Title,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := uc.DB.Session(&gorm.Session{FullSaveAssociations: true}).Create(&newGist)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newGist})
}

func (uc *UserController) CreateCommentOnGist(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.CommentOnGistRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	now := time.Now()

	gistUUID, err := uuid.FromBytes([]byte(payload.GistId))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	newComment := models.Comment{
		GistID:    gistUUID,
		Username:  currentUser.Username,
		Content:   payload.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	result := uc.DB.Create(&newComment)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newComment})
}

// TODO: Fork a gist, remember to check if name is already taken
func (uc *UserController) UpdateUserDetails(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.UpdateUserDetailsRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	userMetadata := currentUser.UserMetadata

	if payload.ProfilePicture != "" {
		userMetadata.ProfilePicture = payload.ProfilePicture
	}
	if payload.Tagline != "" {
		userMetadata.Tagline = payload.Tagline
	}
	if payload.StatusIcon != "" {
		userMetadata.StatusIcon = payload.StatusIcon
	}
	if payload.Location != "" {
		userMetadata.Location = payload.Location
	}
	if payload.Website != "" {
		userMetadata.Website = payload.Website
	}
	if payload.Twitter != "" {
		userMetadata.Twitter = payload.Twitter
	}

	result := uc.DB.Save(&userMetadata)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": userMetadata})
}

func (uc *UserController) UpdateGist(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.UpdateGistRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var gist models.Gist
	result := uc.DB.
		Preload("GistContent").
		First(&gist, "id = ?", payload.GistId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "gist does not exist"})
		return
	}

	if gist.Username != currentUser.Username {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "unauthorized"})
		return
	}

	if payload.Name != "" {
		currentUserGists := currentUser.Gists
		for _, currentUserGist := range currentUserGists {
			if currentUserGist.Name == payload.Name {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Gist with name: '" + payload.Name + "' already exists"})
				return
			}
		}
		gist.Name = payload.Name
	}
	if payload.Title != "" {
		gist.Title = payload.Title
	}
	if payload.Content != "" {
		gist.GistContent.Content = payload.Content
	}
	gist.Private = payload.Private
	gist.UpdatedAt = time.Now()

	result = uc.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&gist)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gist})
}

func (uc *UserController) FollowUser(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	userToFollow := ctx.Params.ByName("userToFollow")

	if currentUser.Username == userToFollow {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "You cannot follow yourself"})
		return
	}

	var userToBeFollowed models.User
	result := uc.DB.Preload("UserMetadata").First(&userToBeFollowed, "username = ?", userToFollow)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user does not exist"})
		return
	}

	// Perform transaction to update both users
	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		// Update current user
		currentUserMetadata := currentUser.UserMetadata
		currentUserMetadata.Following = append(currentUserMetadata.Following, userToBeFollowed.Username)
		result := tx.Save(&currentUserMetadata)
		if result.Error != nil {
			return result.Error
		}

		// Update user to be followed
		userToBeFollowedMetadata := userToBeFollowed.UserMetadata
		userToBeFollowedMetadata.Followers = append(userToBeFollowedMetadata.Followers, currentUser.Username)
		result = tx.Save(&userToBeFollowedMetadata)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "successfully followed user"})
}

func (uc *UserController) UnfollowUser(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	userToUnfollow := ctx.Params.ByName("userToUnfollow")

	if currentUser.Username == userToUnfollow {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "You cannot unfollow yourself"})
		return
	}

	var userToBeUnfollowed models.User
	result := uc.DB.Preload("UserMetadata").First(&userToBeUnfollowed, "username = ?", userToUnfollow)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user does not exist"})
		return
	}

	// Perform transaction to update both users
	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		// Update current user
		currentUserMetadata := currentUser.UserMetadata
		currentUserMetadata.Following = utils.RemoveStringFromSlice(currentUserMetadata.Following, userToBeUnfollowed.Username)
		result := tx.Save(&currentUserMetadata)
		if result.Error != nil {
			return result.Error
		}

		// Update user to be unfollowed
		userToBeUnfollowedMetadata := userToBeUnfollowed.UserMetadata
		userToBeUnfollowedMetadata.Followers = utils.RemoveStringFromSlice(userToBeUnfollowedMetadata.Followers, currentUser.Username)
		result = tx.Save(&userToBeUnfollowedMetadata)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "successfully unfollowed user"})
}

func (uc *UserController) StarGist(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	gistId := ctx.Params.ByName("gistId")

	var gist models.Gist
	result := uc.DB.First(&gist, "id = ?", gistId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "gist does not exist"})
		return
	}

	// Perform transaction to update both users
	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		// Update current user
		currentUserMetadata := currentUser.UserMetadata
		currentUserMetadata.StarredGists = append(currentUserMetadata.StarredGists, gist.ID.String())
		result := tx.Save(&currentUserMetadata)
		if result.Error != nil {
			return result.Error
		}

		// Update gist
		gist.Stars = append(gist.Stars, currentUser.Username)
		result = tx.Save(&gist)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "successfully starred gist"})
}

func (uc *UserController) UnStarGist(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	gistId := ctx.Params.ByName("gistId")

	var gist models.Gist
	result := uc.DB.First(&gist, "id = ?", gistId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "gist does not exist"})
		return
	}

	// Perform transaction to update both users
	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		// Update current user
		currentUserMetadata := currentUser.UserMetadata
		currentUserMetadata.StarredGists = utils.RemoveStringFromSlice(currentUserMetadata.StarredGists, gist.ID.String())
		result := tx.Save(&currentUserMetadata)
		if result.Error != nil {
			return result.Error
		}

		// Update gist
		gist.Stars = utils.RemoveStringFromSlice(gist.Stars, currentUser.Username)
		result = tx.Save(&gist)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "successfully unstarred gist"})
}
