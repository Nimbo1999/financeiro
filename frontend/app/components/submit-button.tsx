import Button, { type ButtonProps } from "@mui/material/Button";

export function SubmitButton({
  type = "submit",
  fullWidth = true,
  size = "large",
  sx = {
    py: 1.5,
    fontSize: "1rem",
    fontWeight: 600,
    background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
    "&:hover": {
      background: "linear-gradient(135deg, #5a6fd8 0%, #6a4391 100%)",
    },
    color: "white",
  },
  variant = "contained",
  children,
  ...props
}: Readonly<ButtonProps>) {
  return (
    <Button
      type={type}
      fullWidth={fullWidth}
      size={size}
      sx={sx}
      variant={variant}
      {...props}
    >
      {children}
    </Button>
  );
}
