// normalizeError converts unknown errors into displayable text.
export function normalizeError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === "string") {
    return error;
  }
  return "エラーが発生しました。";
}
