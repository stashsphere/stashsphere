import { useMemo } from 'react';

export const Pages = ({
  pages,
  currentPage,
  onPageChange,
}: {
  pages: number;
  currentPage: number;
  onPageChange: (page: number) => void;
}) => {
  const allPages = useMemo(() => {
    const pageNumbers = [];
    const lastPage = pages - 1;

    pageNumbers.push(0);

    if (currentPage > 2) {
      pageNumbers.push('...');
    }

    if (currentPage > 1) {
      pageNumbers.push(currentPage - 1);
    }

    if (currentPage != 0) {
      pageNumbers.push(currentPage);
    }

    if (currentPage < lastPage) {
      pageNumbers.push(currentPage + 1);
    }

    if (currentPage < lastPage - 2) {
      pageNumbers.push('...');
    }

    if (lastPage > 1 && currentPage < lastPage - 1) {
      pageNumbers.push(lastPage);
    }

    return pageNumbers;
  }, [currentPage, pages]);

  return (
    <div className="flex justify-center mt-4">
      <ul className="flex list-none">
        {allPages.map((item, index) => (
          <li
            key={index}
            className={`mx-2 ${typeof item === 'number' && item === currentPage ? 'border-2 border-primary' : ''}`}
          >
            {typeof item === 'number' ? (
              <a
                href="#"
                onClick={() => onPageChange(item)}
                className="block py-1 px-3 text-primary hover:bg-primary hover:text-onprimary hover:ring-3"
              >
                {item + 1}
              </a>
            ) : (
              <span className="py-1 px-3 text-gray-500">...</span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};
