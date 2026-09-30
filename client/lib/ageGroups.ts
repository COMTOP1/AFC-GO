/** The age groups the legacy team form offered: "Under 6"…"Under 18", and 19 = "Over 18". */
export const AGE_GROUPS: { value: number; label: string }[] = [
  ...Array.from({ length: 13 }, (_, i) => ({ value: i + 6, label: `Under ${i + 6}` })),
  { value: 19, label: 'Over 18' },
];
