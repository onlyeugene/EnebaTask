"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { login } from "../api";
import { useSignIn } from "../hooks/use-sign-in";
import { nextQuery } from "../next-path";
import { loginSchema, type LoginValues } from "../schemas";
import { showServerError } from "../server-errors";
import { AuthCard } from "./auth-card";
import { FormError, ShowPassword } from "./form-bits";

// login.png, with email + password in place of Eneba's magic link. The
// "Continue with Google / Facebook / Apple" buttons are left out: they are
// real brands, and social login isn't built (PLAN.md section 9).
export function LoginForm({ next }: { next: string }) {
  const signIn = useSignIn();
  const [show, setShow] = useState(false);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ resolver: zodResolver(loginSchema) });

  const onSubmit = async (v: LoginValues) => {
    try {
      await signIn(await login(v.email, v.password), next);
    } catch (e) {
      showServerError(e, setError);
    }
  };

  return (
    <AuthCard title="Log in" question="New user?" linkLabel="Create account" href={`/register${nextQuery(next)}`}>
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex flex-col gap-4">
        <Input
          id="email"
          type="email"
          autoComplete="email"
          placeholder="Email"
          aria-label="Email"
          error={errors.email?.message}
          {...register("email")}
        />
        <Input
          id="password"
          type={show ? "text" : "password"}
          autoComplete="current-password"
          placeholder="Password"
          aria-label="Password"
          error={errors.password?.message}
          {...register("password")}
        />
        <ShowPassword show={show} onToggle={() => setShow((s) => !s)} />
        <FormError message={errors.root?.message} />
        <Button type="submit" disabled={isSubmitting} className="mt-2 w-full text-base">
          {isSubmitting ? "Logging in…" : "Log in"}
        </Button>
      </form>
    </AuthCard>
  );
}
