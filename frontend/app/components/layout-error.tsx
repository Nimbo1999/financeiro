import { CircularProgress } from "@mui/material";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { createPath, Navigate, useAsyncError, useLocation } from "react-router";

export function LayoutError() {
  const error = useAsyncError();
  const location = useLocation();
  const returnPath = createPath(location);

  return (
    <Box
      sx={{
        width: "100vw",
        height: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        gap: 2,
        bgcolor: "background.default",
        p: 3,
      }}
    >
      {error instanceof Error && error.message.includes("401") ? (
        <>
          <Navigate
            to={{
              pathname: "/refresh",
              search: new URLSearchParams({ returnPath }).toString(),
            }}
          />
          <CircularProgress size="6rem" thickness={4.5} />
        </>
      ) : (
        <>
          <Typography variant="h3" color="text.secondary">
            Something went wrong
          </Typography>
          <Typography variant="body1" color="text.primary">
            Please try again later.
          </Typography>
        </>
      )}
    </Box>
  );
}
