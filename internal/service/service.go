package service

type Service struct {
	UserTeamRepository UserTeamRepository
	PR                 PRRepository
}

func NewService(userTeamRepository UserTeamRepository, pr PRRepository) *Service {
	return &Service{
		UserTeamRepository: userTeamRepository,
		PR:                 pr,
	}
}
