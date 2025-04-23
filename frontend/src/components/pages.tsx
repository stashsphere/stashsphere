export const Pages = ({
  pages,
  currentPage,
  onPageChange,
}: {
  pages: number;
  currentPage: number;
  onPageChange: (page: number) => void;
}) => {
  const allPages = Array.from(Array(pages === 0 ? 1 : pages).keys());
  return (
    <div className="flex justify-center mt-4">
      <ul className="flex list-none">
        {allPages.map((page) => (
          <li
            key={page + 1}
            className={`mx-2 ${page === currentPage ? 'border-2 border-primary' : ''}`}
          >
            <a
              href="#"
              onClick={() => onPageChange(page)}
              className="block py-1 px-3 text-primary hover:bg-primary hover:text-onprimary hover:ring-3"
            >
              {page + 1}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
};
