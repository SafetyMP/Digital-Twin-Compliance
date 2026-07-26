import type { ReactNode } from "react";

type Variant = "loading" | "empty" | "error";

const styles: Record<Variant, string> = {
  loading: "border-slate-700 bg-slate-900/60 text-slate-300",
  empty: "border-slate-800 bg-slate-900/40 text-slate-400",
  error: "border-red-800 bg-red-950/40 text-red-200",
};

export function StatusMessage({
  variant,
  children,
  action,
}: {
  variant: Variant;
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div
      role={variant === "error" ? "alert" : "status"}
      className={`rounded-lg border p-4 text-sm ${styles[variant]}`}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>{children}</div>
        {action}
      </div>
    </div>
  );
}

export function LoadingState({ label = "Loading…" }: { label?: string }) {
  return <StatusMessage variant="loading">{label}</StatusMessage>;
}

export function EmptyState({
  children,
  action,
}: {
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <StatusMessage variant="empty" action={action}>
      {children}
    </StatusMessage>
  );
}

export function ErrorState({
  children,
  action,
}: {
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <StatusMessage variant="error" action={action}>
      {children}
    </StatusMessage>
  );
}
