import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import { authApi } from "../api/endpoints";
import { errText } from "../api/client";
import { applyAuthResponse } from "../stores/auth";
import { useI18n } from "../i18n";

const schema = z.object({
  username: z.string().min(2),
  email: z.string().email(),
  password: z.string().min(6),
  display_name: z.string().optional(),
});
type FormData = z.infer<typeof schema>;

export default function RegisterPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    try {
      const res = await authApi.register(data);
      applyAuthResponse(res);
      nav("/");
    } catch (e) {
      setError("username", { message: errText(e) });
    }
  };

  return (
    <div className="flex h-full items-center justify-center bg-bg p-6">
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="w-full max-w-sm space-y-4 rounded-2xl border border-line bg-surface p-6"
      >
        <h1 className="text-center text-xl font-bold">{t("action.register")}</h1>

        <label className="block text-sm">
          {t("auth.username")}
          <input {...register("username")} className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2" />
        </label>
        {errors.username && <p className="text-sm text-danger">{errors.username.message}</p>}

        <label className="block text-sm">
          {t("auth.email")}
          <input type="email" {...register("email")} className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2" />
        </label>
        {errors.email && <p className="text-sm text-danger">{errors.email.message}</p>}

        <label className="block text-sm">
          {t("auth.password")}
          <input
            type="password"
            {...register("password")}
            className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2"
            autoComplete="new-password"
          />
        </label>
        {errors.password && <p className="text-sm text-danger">{errors.password.message}</p>}

        <label className="block text-sm">
          {t("auth.displayName")}
          <input {...register("display_name")} className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2" />
        </label>

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-primary px-4 py-2.5 font-semibold text-primary-ink disabled:opacity-50"
        >
          {t("action.register")}
        </button>
        <button type="button" onClick={() => nav("/login")} className="w-full text-sm text-primary underline">
          {t("auth.toLogin")}
        </button>
      </form>
    </div>
  );
}
