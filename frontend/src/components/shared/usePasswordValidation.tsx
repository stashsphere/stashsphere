export const usePasswordValidation = (
  password: string,
  confirmPassword: string,
  minLength: number = 8
) => {
  const isValid = password.length >= minLength && password === confirmPassword;
  return { isValid };
};
