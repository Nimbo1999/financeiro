import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export interface MaintanceBoxProps {
  description?: string;
}

export function MaintanceBox({ description }: Readonly<MaintanceBoxProps>) {
  return (
    <Box
      sx={{
        mt: 4,
        p: 8,
        textAlign: "center",
        backgroundColor: "background.paper",
        borderRadius: 3,
        border: "2px dashed",
        borderColor: "divider",
      }}
    >
      <Typography variant="h4" color="text.secondary" gutterBottom>
        🔧👷🏻 Under Construction 👨‍🔧🔨
      </Typography>
      <Typography variant="body2" color="text.primary">
        {description ?? "A new feature is being developed here."}
      </Typography>
    </Box>
  );
}
