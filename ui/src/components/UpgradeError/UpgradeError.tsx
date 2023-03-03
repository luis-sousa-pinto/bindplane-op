import { gql } from "@apollo/client";
import { Alert, AlertTitle } from "@mui/material";
import { useClearAgentUpgradeErrorMutation } from "../../graphql/generated";

interface UpgradeErrorProps {
  agentId: string;
  upgradeError?: string | null;
  onClearSuccess: () => void;
  onClearFailure: () => void;
}

gql`
  mutation ClearAgentUpgradeError($input: ClearAgentUpgradeErrorInput!) {
    clearAgentUpgradeError(input: $input)
  }
`;

/**
 * UpgradeError displays the agent upgrade error message and can clear it with
 * the clearAgentUpgradeError mutation.  It calls onClearSuccess when the
 * mutation is successful and onClearFailure when it fails.
 */
export const UpgradeError: React.FC<UpgradeErrorProps> = ({
  agentId,
  upgradeError,
  onClearFailure,
  onClearSuccess,
}) => {
  const [clearAgentUpgradeError] = useClearAgentUpgradeErrorMutation({
    variables: { input: { agentId } },
    onCompleted: onClearSuccess,
    onError: onClearFailure,
  });

  async function handleClearError() {
    await clearAgentUpgradeError();
  }

  if (upgradeError == null) {
    return null;
  }

  return (
    <Alert severity="error" onClose={handleClearError}>
      <AlertTitle>Upgrade Error</AlertTitle>
      {upgradeError}
    </Alert>
  );
};
