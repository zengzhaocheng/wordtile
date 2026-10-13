import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import { authApi } from "../api/endpoints";
import { errText } from "../api/client";
import { applyAuthResponse } from "../stores/auth";
import { useI18n } from "../i18n";

const schema = z.object({
  identifier: z.string().min(1),
  password: z.string().min(1),
});
type FormData = z.infer<typeof schema>;

export default function LoginPage() {
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
      const res = await authApi.login(data);
      applyAuthResponse(res);
      nav("/");
    } catch (e) {
      setError("password", { message: errText(e) });
    }
  };

  return (
    <div className="flex h-full items-center justify-center bg-bg p-6">
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="w-full max-w-sm space-y-4 rounded-2xl border border-line bg-surface p-6"
      >
        <h1 className="text-center text-2xl font-bold">🀄 WordTile</h1>
        <p className="text-center text-sm text-muted">{t("app.tagline")}</p>

        <label className="block text-sm">
          {t("auth.identifier")}
          <input
            {...register("identifier")}
            className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2"
            autoComplete="username"
          />
        </label>
        <label className="block text-sm">
          {t("auth.password")}
          <input
            type="password"
            {...register("password")}
            className="mt-1 w-full rounded-lg border border-line bg-bg px-3 py-2"
            autoComplete="current-password"
          />
        </label>
        {errors.password && <p className="text-sm text-danger">{errors.password.message}</p>}
        <p className="text-xs text-muted">演示账号：demo / demo1234</p>

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-primary px-4 py-2.5 font-semibold text-primary-ink disabled:opacity-50"
        >
          {t("action.login")}
        </button>
        <button type="button" onClick={() => nav("/register")} className="w-full text-sm text-primary underline">
          {t("auth.toRegister")}
        </button>
      </form>
    </div>
  );
}
