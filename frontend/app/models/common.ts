export interface ListReponse<Entity> {
  data: Entity[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
