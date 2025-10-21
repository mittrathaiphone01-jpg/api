package service

import (
	"fmt"
	"rrmobile/config"
	"rrmobile/respository"

	"time"

	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

type usersService struct {
	usersRepository respository.UserRepository
}

func NewUsersService(usersRepository respository.UserRepository) UsersService {
	return &usersService{usersRepository: usersRepository}
}
func (s usersService) GetAllUsers(usernames []string, date string, limit, offset int) ([]UserResponseAll, error) {
	users, err := s.usersRepository.GetAllUsers(usernames, date, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: ")
	}

	var usersResponse []UserResponseAll
	for _, user := range users {
		var isActive bool
		if user.Is_active != nil {
			isActive = *user.Is_active
		} else {
			isActive = false // หรือ default true แล้วแต่ business logic
		}

		usersResponse = append(usersResponse, UserResponseAll{
			Id:        user.Id,
			Username:  user.Username,
			FullName:  user.FullName,
			RoleID:    int(user.RoleID),
			Is_active: isActive,
		})
	}


	return usersResponse, nil
}
func (s usersService) CountUsers(usernames []string, date string) (int, error) {
	var count int64
	err := s.usersRepository.CountUsers(usernames, date, &count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users")
	}
	return int(count), nil
}

func (s *usersService) GetUserById(id uint) (*UserResponse, error) {
	user, err := s.usersRepository.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	var isActive bool
	if user.Is_active != nil {
		isActive = *user.Is_active
	} else {
		isActive = false // หรือ default true แล้วแต่ business logic
	}
	response := UserResponse{
		Id:        user.Id,
		Username:  user.Username,
		FullName:  user.FullName,
		RoleID:    int(user.RoleID),
		CreatedAt: user.CreatedAt,
		Is_active: isActive,
	}
	return &response, nil

}
func (s *usersService) CreateUser(request NewUserRequest) (*UserResponse, error) {
	isActive := true

	user := respository.User{
		FullName:  request.FullName,
		Password:  request.Password,
		Username:  request.Username,
		CreatedAt: time.Now(),
		RoleID:    2,
		Is_active: &isActive,
	}

	validate := validator.New()
	err := validate.Var(user.FullName, "required")
	if err != nil {
		return nil, fmt.Errorf("Please enter a  FullName")
	}
	err = validate.Var(user.Username, "required")
	if err != nil {
		return nil, fmt.Errorf("Please enter a  Username")
	}
	err = validate.Var(request, "required")
	if err != nil {
		return nil, fmt.Errorf("Insert Cannot emtpy")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password")
	}
	user.Password = string(passwordHash) // Store hashed password
	newUser, err := s.usersRepository.AddUser(user)
	if err != nil {
		return nil, fmt.Errorf("Failed to create user")
	}
	createdAtInThai, err := config.ConvertToThaiTime(newUser.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert time to Thai format")
	}
	response := UserResponse{
		Id:        newUser.Id,
		Username:  newUser.Username,
		FullName:  newUser.FullName,
		CreatedAt: createdAtInThai,
		RoleID:    int(newUser.RoleID),
		Is_active: *newUser.Is_active,
	}
	return &response, nil
}

func (s *usersService) EditUser(id uint, newUsername string, newFullName string, newPassword string, newActive *bool) (*UserResponse, error) {
	user, err := s.usersRepository.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID %d not found", id)
	}

	user.FullName = newFullName
	user.Username = newUsername

	if newPassword != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("Failed to hash password")
		}
		user.Password = string(passwordHash)
	}

	if newActive != nil {
		user.Is_active = newActive // assign pointer ตรง ๆ
	}

	updatedUser, err := s.usersRepository.UpdateUser(id, *user)
	if err != nil {
		return nil, fmt.Errorf("Failed to update user: %w", err)
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	thaiTime := updatedUser.UpdatedAt.In(loc)

	return &UserResponse{
		Id:        updatedUser.Id,
		Username:  updatedUser.Username,
		FullName:  updatedUser.FullName,
		Is_active: *updatedUser.Is_active,
		UpdatedAt: thaiTime,
	}, nil
}

func (s *usersService) DeleteUser(id uint) error {
	return s.usersRepository.DeleteUser(id)
}

func (s *usersService) ResetPasswordUser(id uint, newPassword string) (*UserResponse, error) {
	user, err := s.usersRepository.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("User with ID not found", id)
	}
	// Hash the new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password")
	}
	user.Password = string(passwordHash) // Store the hashed password

	updatedUser, err := s.usersRepository.UpdateUser(id, *user)
	if err != nil {
		return nil, fmt.Errorf("Failed to update user password")
	}
	response := UserResponse{
		Id:        updatedUser.Id,
		Username:  updatedUser.Username,
		FullName:  updatedUser.FullName,
		CreatedAt: updatedUser.CreatedAt,
	}
	return &response, nil
}
