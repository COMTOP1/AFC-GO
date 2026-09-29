import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Table, TBody, Td, Th, THead, Tr } from './Table';

describe('Table', () => {
  it('renders an accessible table inside a horizontal scroller', () => {
    const { container } = render(
      <Table>
        <THead>
          <Tr>
            <Th>Name</Th>
            <Th>Role</Th>
          </Tr>
        </THead>
        <TBody>
          <Tr>
            <Td>Jo Smith</Td>
            <Td>Manager</Td>
          </Tr>
        </TBody>
      </Table>,
    );
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getAllByRole('columnheader').map((th) => th.textContent)).toEqual([
      'Name',
      'Role',
    ]);
    expect(screen.getByRole('cell', { name: 'Jo Smith' })).toBeInTheDocument();
    expect(container.firstElementChild).toHaveClass('overflow-x-auto');
  });
});
