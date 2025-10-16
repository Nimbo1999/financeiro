import { CircularProgress } from "@mui/material";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import {
  createPath,
  useAsyncError,
  useLocation,
  useNavigate,
} from "react-router";
import { useEffect } from "react";

export function LayoutError() {
  const error = useAsyncError();
  const location = useLocation();
  const navigate = useNavigate();
  const returnPath = createPath(location);

  useEffect(() => {
    const timer = setTimeout(() => {
      navigate({
        pathname: "/refresh",
        search: new URLSearchParams({ returnPath }).toString(),
      });
    }, 1000);
    return () => clearTimeout(timer);
  }, [error, navigate, returnPath]);

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
      <Typography variant="h3" color="text.secondary">
        Verifying your session...
      </Typography>

      <Typography variant="body1" color="text.primary">
        Please wait until we redirect you.
      </Typography>

      <CircularProgress size="6rem" thickness={4.5} />
    </Box>
  );
}
