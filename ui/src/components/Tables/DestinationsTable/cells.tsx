import { gql } from "@apollo/client";
import { useGetDestinationTypeDisplayInfoQuery } from "../../../graphql/generated";

import styles from "./cells.module.scss";

interface ResourceTypeCellProps {
  type: string;
  icon?: boolean;
}

export const DestinationTypeCell: React.FC<ResourceTypeCellProps> = ({
  type,
  icon,
}) => {
  const { data } = useGetDestinationTypeDisplayInfoQuery({
    variables: { name: type },
  });
  return data?.destinationType ? (
    <div className={styles.cell}>
      <span
        className={styles.icon}
        style={{
          backgroundImage: `url(${data.destinationType?.metadata.icon ?? ""})`,
        }}
      />
      {!icon && data.destinationType?.metadata.displayName}
    </div>
  ) : (
    <div>{type}</div>
  );
};

gql`
  query getDestinationTypeDisplayInfo($name: String!) {
    destinationType(name: $name) {
      metadata {
        id
        name
        version
        displayName
        icon
      }
    }
  }
`;

gql`
  query getSourceTypeDisplayInfo($name: String!) {
    sourceType(name: $name) {
      metadata {
        id
        name
        version
        displayName
        icon
      }
    }
  }
`;
