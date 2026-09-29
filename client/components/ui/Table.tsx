import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export function Table({ className, ...props }: ComponentProps<'table'>) {
  return (
    <div className="w-full overflow-x-auto">
      <table className={clsx('w-full border-collapse text-sm', className)} {...props} />
    </div>
  );
}

export function THead(props: ComponentProps<'thead'>) {
  return <thead {...props} />;
}

export function TBody({ className, ...props }: ComponentProps<'tbody'>) {
  return <tbody className={clsx('[&>tr:hover]:bg-surface', className)} {...props} />;
}

export function Tr(props: ComponentProps<'tr'>) {
  return <tr {...props} />;
}

export function Th({ className, ...props }: ComponentProps<'th'>) {
  return (
    <th
      className={clsx(
        'border-b-2 border-line px-3 py-2 text-left text-[11px] font-bold tracking-wider text-muted uppercase',
        className,
      )}
      {...props}
    />
  );
}

export function Td({ className, ...props }: ComponentProps<'td'>) {
  return <td className={clsx('border-b border-line px-3 py-2.5', className)} {...props} />;
}
