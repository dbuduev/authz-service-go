package core

type AuthorisationCore struct {
	repository Repository
}

func CreateAuthorisationCore(repository Repository) AuthorisationCore {
	return AuthorisationCore{repository: repository}
}
