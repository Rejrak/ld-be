package common

type UserAccess struct {
	List   bool
	Detail bool
	Edit   bool
}

func NewUserAccess() *UserAccess {
	return new(UserAccess)
}
