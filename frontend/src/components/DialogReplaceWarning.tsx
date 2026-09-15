// Shared inline "this will replace an existing item" warning banner, used by
// both the Filtres and Sources M3U create dialogs instead of a native
// confirm(). See the Create Filter Configuration and M3U Sources Management
// requirements' shared trigger/dialog/warning-banner presentation.
export function DialogReplaceWarning({ message }: { message: string }) {
  return (
    <div className="dialog-warning-banner" role="alert">
      ⚠️ {message}
    </div>
  );
}
