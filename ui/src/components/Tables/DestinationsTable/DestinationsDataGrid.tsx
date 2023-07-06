import { Stack } from "@mui/material";
import {
  DataGrid,
  DataGridProps,
  GridCellParams,
  GridColDef,
  GridRowSelectionModel,
  GridValueGetterParams,
} from "@mui/x-data-grid";
import { isFunction } from "lodash";
import { memo } from "react";
import { DestinationTypeCell } from "./cells";

import styles from "./cells.module.scss";

export enum DestinationsTableField {
  NAME = "name",
  TYPE = "type",
  ICON_AND_NAME = "icon",
}

interface DestinationsDataGridProps extends Omit<DataGridProps, "columns"> {
  setSelectionModel?: (names: GridRowSelectionModel) => void;
  onEditDestination: (name: string) => void;
  loading: boolean;
  columnFields?: DestinationsTableField[];
  minHeight?: string;
  maxHeight?: string;
  selectionModel?: GridRowSelectionModel;
  destinationsPage?: boolean;
  allowSelection: boolean;
}

export const DestinationsDataGrid: React.FC<DestinationsDataGridProps> = memo(
  ({
    setSelectionModel,
    onEditDestination,
    columnFields,
    minHeight,
    maxHeight,
    selectionModel,
    destinationsPage,
    allowSelection,
    ...dataGridProps
  }) => {
    function renderNameCell(
      cellParams: GridCellParams<any, string>
    ): JSX.Element {
      if (cellParams.row.kind === "Destination") {
        return (
          <button
            onClick={() => onEditDestination(cellParams.value!)}
            className={styles.link}
          >
            {cellParams.value}
          </button>
        );
      }

      return renderStringCell(cellParams);
    }

    function renderNameAndIconCell(
      cellParams: GridCellParams<any, { name: string; type: string }>
    ): JSX.Element {
      return (
        <>
          <DestinationTypeCell icon type={cellParams?.value?.type ?? ""} />
          <button
            onClick={() => onEditDestination(cellParams.value?.name!)}
            className={styles.link}
          >
            {cellParams.value?.name}
          </button>
        </>
      );
    }

    const columns: GridColDef[] = (columnFields || []).map((field) => {
      switch (field) {
        case DestinationsTableField.NAME:
          return {
            field: DestinationsTableField.NAME,
            width: 300,

            headerName: "Name",
            valueGetter: (params: GridValueGetterParams) =>
              params.row.metadata.name,
            renderCell: renderNameCell,
          };
        case DestinationsTableField.TYPE:
          return {
            field: DestinationsTableField.TYPE,
            flex: 1,
            headerName: "Type",
            valueGetter: (params: GridValueGetterParams) =>
              params.row.spec.type,
            renderCell: renderTypeCell,
          };
        case DestinationsTableField.ICON_AND_NAME:
          return {
            field: DestinationsTableField.ICON_AND_NAME,
            flex: 1,
            headerName: "Name",
            valueGetter: (params: GridValueGetterParams) => {
              return {
                type: params.row.spec.type,
                name: params.row.metadata.name,
              };
            },
            sortComparator: (v1, v2: { name: string; type: string }) => {
              return v1.name.localeCompare(v2.name);
            },
            renderCell: renderNameAndIconCell,
          };
        default:
          return { field: DestinationsTableField.TYPE };
      }
    });

    return (
      <DataGrid
        {...dataGridProps}
        checkboxSelection={isFunction(setSelectionModel) && allowSelection}
        onRowSelectionModelChange={setSelectionModel}
        components={{
          NoRowsOverlay: () => (
            <Stack height="100%" alignItems="center" justifyContent="center">
              No Destinations
            </Stack>
          ),
        }}
        style={{ minHeight, maxHeight }}
        disableRowSelectionOnClick
        getRowId={(row) => `${row.kind}|${row.metadata.name}`}
        columns={columns}
        rowSelectionModel={selectionModel}
      />
    );
  }
);

function renderTypeCell(cellParams: GridCellParams<any, string>): JSX.Element {
  return <DestinationTypeCell type={cellParams.value ?? ""} />;
}

function renderStringCell(
  cellParams: GridCellParams<any, string>
): JSX.Element {
  return <>{cellParams.value}</>;
}

DestinationsDataGrid.defaultProps = {
  minHeight: "calc(100vh - 250px)",
  columnFields: [DestinationsTableField.NAME, DestinationsTableField.TYPE],
};
