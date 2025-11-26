package mapper

import (
	"pr-service/internal/domain"
	"pr-service/internal/handlers/dto"
)

func UserToDTO(user domain.User, team string) dto.UserDTO {
	return dto.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		IsActive: user.IsActive,
		TeamName: team,
	}
}

func UserFromDTO(userDTO dto.UserDTO) domain.User {
	return domain.User{
		ID:       userDTO.ID,
		Username: userDTO.Username,
		IsActive: userDTO.IsActive,
		TeamName: userDTO.TeamName,
	}
}

func UsersToDTO(users []domain.User, team string) []dto.UserDTO {
	result := make([]dto.UserDTO, len(users))
	for i, user := range users {
		result[i] = UserToDTO(user, team)
	}
	return result
}

func UsersFromDTO(userDTOs []dto.UserDTO) []domain.User {
	result := make([]domain.User, len(userDTOs))
	for i, userDTO := range userDTOs {
		result[i] = UserFromDTO(userDTO)
	}
	return result
}
