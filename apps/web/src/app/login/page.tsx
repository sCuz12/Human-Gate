import { AuthCard } from "@/components/auth/auth-card";
import { LoginForm } from "@/components/auth/login-form";

export default function LoginPage() {
  return (
    <AuthCard
      alternateHref="/register"
      alternateLabel="Need an account?"
      alternateText="Register"
      description="Sign in with your Supabase account to access workspace and API key management."
      eyebrow="Sign in"
      title="Welcome back."
    >
      <LoginForm />
    </AuthCard>
  );
}
