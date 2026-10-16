import { ReactNode } from "react";

interface PageHeaderProps {
  title: string;
  description: string;
  actions?: ReactNode;
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800/80">
      <div>
        <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white">{title}</h1>
        <p className="mt-1 text-xs sm:text-sm text-zinc-400 leading-relaxed max-w-3xl">{description}</p>
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  );
}
