import { useState } from "react";
import { Button } from "@/components/ui/button";

interface RestartToApplyButtonProps {
  onRestart: () => Promise<void>;
}

/**
 * The action on the "restart to apply" toast. It disables itself once pressed,
 * so a second click while the first restart is under way does not queue another.
 */
export function RestartToApplyButton({ onRestart }: RestartToApplyButtonProps) {
  const [isRestarting, setIsRestarting] = useState(false);

  const handleClick = async () => {
    setIsRestarting(true);
    await onRestart();
  };

  return (
    <Button size="sm" variant="outline" disabled={isRestarting} onClick={handleClick}>
      {isRestarting ? "Restarting..." : "Restart now"}
    </Button>
  );
}
