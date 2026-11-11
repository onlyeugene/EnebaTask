"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { register as createAccount } from "../api";
import { useSignIn } from "../hooks/use-sign-in";
import { nextQuery } from "../next-path";
import { registerSchema, type RegisterValues } from "../schemas";
import { showServerError } from "../server-errors";
import { AuthCard } from "./auth-card";
import { FormError, ShowPassword } from "./form-bits";
import { PasswordRules } from "./password-rules";

// signup.png. Left out: the offers checkbox (no newsletter), the social
// buttons (real brands, not built), and the terms line, replaced by a demo note.
export function RegisterForm({ next }: { next: string }) {
  const signIn = useSignIn();
  const [show, setShow] = useState(false);
  const {
    register,
    handleSubmit,
    setError,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({ resolver: zodResolver(registerSchema), defaultValues: { password: "" } });

  const onSubmit = async (v: RegisterValues) => {
    try {
      await signIn(await createAccount(v.email, v.password), next);
    } catch (e) {
      showServerError(e, setError);
    }
  };

  return (
    <AuthCard title="Create account" question="Already have an account?" linkLabel="Login" href={`/login${nextQuery(next)}`}>
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
          autoComplete="new-password"
          placeholder="Password"
          aria-label="Password"
          error={errors.password?.message}
          {...register("password")}
        />
        <Input
          id="repeat"
          type={show ? "text" : "password"}
          autoComplete="new-password"
          placeholder="Repeat password"
          aria-label="Repeat password"
          error={errors.repeat?.message}
          {...register("repeat")}
        />
        <div className="flex flex-col">
          <ShowPassword show={show} onToggle={() => setShow((s) => !s)} />
          {/* watch() re-renders on each key, so the rules tick live. */}
          <PasswordRules password={watch("password")} />
        </div>
        <FormError message={errors.root?.message} />
        <Button type="submit" disabled={isSubmitting} className="mt-2 w-full text-base">
          {isSubmitting ? "Creating account…" : "Create account"}
        </Button>
        <p className="text-xs text-text-secondary">
          Demo project: no emails are sent, so any address you can remember works.
        </p>
      </form>
    </AuthCard>
  );
}
