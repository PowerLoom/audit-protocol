from skydb_utils import SkydbTable
import random
import string

table_name = ''.join([random.choice(string.ascii_letters) for i in range(20)])

table = SkydbTable(table_name=table_name,
					columns=['c1','c2'],
					seed='Seed of random characters'
				)

row_index = table.add_row({'c1':'ROW-1'})
row_index = table.add_row({'c1':'ROW-2'})
row_index = table.add_row({'c1':'ROW-3'})
row_index = table.add_row({'c1':'ROW-4'})
row_index = table.add_row({'c1':'ROW-5'})
row_index = table.add_row({'c1':'ROW-1'})
row_index = table.add_row({'c1':'ROW-2'})
row_index = table.add_row({'c1':'ROW-3'})
row_index = table.add_row({'c1':'ROW-4'})
row_index = table.add_row({'c1':'ROW-5'})

print(table.fetch({'c1':'ROW-5'}, row_index,  n_rows=2))

