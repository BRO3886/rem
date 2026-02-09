#ifndef EVENTKIT_DARWIN_H
#define EVENTKIT_DARWIN_H

// ek_fetch_lists returns a JSON array of reminder lists.
// Caller must free the returned string with ek_free.
// Returns NULL on error (error written to ek_last_error).
char* ek_fetch_lists(void);

// ek_fetch_reminders returns a JSON array of reminders matching the given filters.
// All filter parameters may be NULL to skip that filter.
// Caller must free the returned string with ek_free.
char* ek_fetch_reminders(const char* list_name,
                         const char* completed_filter,
                         const char* search_query,
                         const char* due_before,
                         const char* due_after);

// ek_get_reminder returns a single reminder as JSON by ID or ID prefix.
// Caller must free the returned string with ek_free.
// Returns NULL if not found.
char* ek_get_reminder(const char* target_id);

// ek_free frees a string returned by the above functions.
void ek_free(char* ptr);

// ek_last_error returns the last error message, or NULL if no error.
// The returned string is valid until the next call to any ek_ function.
const char* ek_last_error(void);

#endif
