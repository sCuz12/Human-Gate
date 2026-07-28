import { AuthCard } from "@/components/auth/auth-card";
import { RegisterForm } from "@/components/auth/register-form";

export default function RegisterPage() {
  return (
    <AuthCard
      alternateHref="/login"
      alternateLabel="Already have an account?"
      alternateText="Sign in"
      description="Create your account first. The database trigger will sync your auth user into the application users table."
      eyebrow="Register"
      title="Start your workspace."
    >
      <RegisterForm />
    </AuthCard>
  );
}
