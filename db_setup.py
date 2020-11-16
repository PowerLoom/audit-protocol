import sqlite3

sqlite_conn = sqlite3.connect('auditprotocol_1.db')
sqlite_cursor = sqlite_conn.cursor()

try:
    sqlite_cursor.execute('''CREATE TABLE api_keys 
    (
        token text, 
        apiKey text
    )
''')
except sqlite3.OperationalError:
    pass


try:
    sqlite_cursor.execute('''CREATE TABLE accounting_records 
    (
        token text, 
        cid text, 
        localCID text, 
        txHash text, 
        confirmed integer,
        timestamp integer
    )
''')
except sqlite3.OperationalError:
    pass


try:
    sqlite_cursor.execute('''CREATE TABLE retrievals_single 
    (
        requestID text,
        cid text,
        localCID text,
        retrievedFile text, 
        completed integer
    )
''')
except sqlite3.OperationalError:
    pass


try:
    sqlite_cursor.execute('''
        CREATE TABLE retrievals_bulk
        (
            requestID text,
            api_key text,
            token text,
            retrievedFile text,
            completed integer
        )
    ''')
except sqlite3.OperationalError:
    pass

try:
    sqlite_conn.commit()
except:
    pass

try:
    sqlite_cursor.close()
    sqlite_conn.close()
except:
    pass