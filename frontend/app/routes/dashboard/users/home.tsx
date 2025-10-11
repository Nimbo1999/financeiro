import Container from "@mui/material/Container";
import InputAdornment from "@mui/material/InputAdornment";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TablePagination from "@mui/material/TablePagination";
import TableSortLabel from "@mui/material/TableSortLabel";
import TextField from "@mui/material/TextField";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import SearchIcon from "@mui/icons-material/Search";
import AddCircleIcon from "@mui/icons-material/AddCircle";
import { useSearchParams, createSearchParams, Await, Link } from "react-router";
import { Suspense, useCallback, useMemo } from "react";

import { getCurrentSession } from "~/session";
import type { User } from "~/models/user";

import type { Route } from "./+types/home";
import { FinanceiroUserService } from "~/services/user.service";
import { FetchClient } from "~/clients/http";
import { env } from "~/environment";
import type { ListReponse } from "~/models/common";
import { Button } from "@mui/material";

function getSearchParams(searchParams: URLSearchParams): URLSearchParams {
  const params = createSearchParams(searchParams);
  if (params.has("page") === false) {
    params.set("page", "1");
  }
  if (params.has("page_size") === false) {
    params.set("page_size", "10");
  }
  if (params.has("order_by") === false) {
    params.set("order_by", "updated_at");
  }
  if (params.has("sort") === false) {
    params.set("sort", "desc");
  }
  return params;
}

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url);
  const searchParams = getSearchParams(url.searchParams);
  const { accessToken } = await getCurrentSession(
    request.headers.get("Cookie")
  );

  const service = new FinanceiroUserService(
    FetchClient.new(env.API_BASE_URL, accessToken)
  );
  const usersPromise = service.getList(searchParams);

  return { searchParams: Object.fromEntries(searchParams), usersPromise };
}

interface UsersTableProps {
  users: ListReponse<User>;
  searchParams: Record<string, string>;
}

function UsersTable({ users, searchParams }: UsersTableProps) {
  const [, setSearchParams] = useSearchParams();

  const page = parseInt(searchParams.page || "1", 10) - 1; // MUI uses 0-based indexing
  const pageSize = parseInt(searchParams.page_size || "10", 10);
  const orderBy = searchParams.order_by || "updated_at";
  const sort = (searchParams.sort || "desc") as "asc" | "desc";

  const updateSearchParams = useCallback(
    (updates: Record<string, string>) => {
      const newParams = new URLSearchParams(searchParams);
      Object.entries(updates).forEach(([key, value]) => {
        if (value) {
          newParams.set(key, value);
        } else {
          newParams.delete(key);
        }
      });
      setSearchParams(newParams, { replace: true });
    },
    [setSearchParams, searchParams]
  );

  const handleChangePage = useCallback(
    (_event: unknown, newPage: number) => {
      updateSearchParams({ page: (newPage + 1).toString() });
    },
    [updateSearchParams]
  );

  const handleChangeRowsPerPage = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      updateSearchParams({
        page_size: event.target.value,
        page: "1", // Reset to first page
      });
    },
    [updateSearchParams]
  );

  const handleSort = useCallback(
    (column: string) => {
      const isAsc = orderBy === column && sort === "asc";
      updateSearchParams({
        order_by: column,
        sort: isAsc ? "desc" : "asc",
      });
    },
    [orderBy, sort, updateSearchParams]
  );

  const handleSearchSubmit = useCallback(
    (event: React.FormEvent) => {
      event.preventDefault();
      const formData = new FormData(event.target as HTMLFormElement);
      updateSearchParams({
        ...Object.fromEntries(formData.entries()),
        page: "1", // Reset to first page on new search
      });
    },
    [updateSearchParams]
  );

  const columns = useMemo(
    () => [
      { id: "full_name", label: "Full Name", sortable: true },
      { id: "email", label: "Email", sortable: true },
      { id: "created_at", label: "Created At", sortable: true },
      { id: "updated_at", label: "Updated At", sortable: true },
    ],
    []
  );

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Box>
      <Box
        sx={{
          mb: 3,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <Box
          component="form"
          onSubmit={handleSearchSubmit}
          sx={{ display: "flex", gap: 1 }}
        >
          <TextField
            size="small"
            placeholder="Search by full name"
            type="search"
            name="full_name"
            id="full_name"
            sx={{ minWidth: 250 }}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment
                    position="start"
                    sx={{ color: "text.secondary" }}
                  >
                    <SearchIcon color="inherit" />
                  </InputAdornment>
                ),
              },
            }}
          />
        </Box>

        <Button
          variant="contained"
          color="primary"
          startIcon={<AddCircleIcon />}
          component={Link}
          to="create"
        >
          Create user
        </Button>
      </Box>

      <Paper>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                {columns.map((column) => (
                  <TableCell key={column.id}>
                    {column.sortable ? (
                      <TableSortLabel
                        active={orderBy === column.id}
                        direction={orderBy === column.id ? sort : "asc"}
                        onClick={() => handleSort(column.id)}
                      >
                        {column.label}
                      </TableSortLabel>
                    ) : (
                      column.label
                    )}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {users.data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={columns.length} align="center">
                    <Typography
                      variant="body2"
                      color="text.secondary"
                      sx={{ py: 3 }}
                    >
                      No users found
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                users.data.map((user) => (
                  <TableRow key={user.id} hover>
                    <TableCell>{user.full_name}</TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>{formatDate(user.created_at)}</TableCell>
                    <TableCell>{formatDate(user.updated_at)}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>

        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50, 100]}
          component="div"
          count={users.total}
          rowsPerPage={pageSize}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
    </Box>
  );
}

export default function UserList({
  loaderData: { searchParams, usersPromise },
}: Route.ComponentProps) {
  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      <Suspense
        fallback={
          <Box
            sx={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              minHeight: 400,
            }}
          >
            <CircularProgress />
          </Box>
        }
      >
        <Await resolve={usersPromise}>
          {(users) => <UsersTable users={users} searchParams={searchParams} />}
        </Await>
      </Suspense>
    </Container>
  );
}

export function meta() {
  return [
    { title: "Finance tracker - Users" },
    {
      name: "description",
      content: "This is where you can manage your users",
    },
  ];
}
