package auth

// PasswordConfig holds the password hashing parameters (legacy PBKDF2
// iterations plus scrypt parameters), loaded from the environment.
type PasswordConfig struct {
	Iterations              int
	ScryptWorkFactor        int
	ScryptBlockSize         int
	ScryptParallelismFactor int
	KeyLength               int
}
