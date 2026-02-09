#import <EventKit/EventKit.h>
#import <Foundation/Foundation.h>
#include "eventkit_darwin.h"
#include <stdlib.h>
#include <string.h>

// Thread-local error message
static __thread char* last_error = NULL;

static void set_error(NSString* msg) {
    if (last_error) {
        free(last_error);
        last_error = NULL;
    }
    if (msg) {
        last_error = strdup([msg UTF8String]);
    }
}

const char* ek_last_error(void) {
    return last_error;
}

void ek_free(char* ptr) {
    if (ptr) free(ptr);
}

// Shared store instance with access request
static EKEventStore* get_store(void) {
    static EKEventStore* store = nil;
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        store = [[EKEventStore alloc] init];
        // Request access to reminders (triggers TCC prompt on first use)
        dispatch_semaphore_t sem = dispatch_semaphore_create(0);
        if (@available(macOS 14.0, *)) {
            [store requestFullAccessToRemindersWithCompletion:^(BOOL granted, NSError* error) {
                dispatch_semaphore_signal(sem);
            }];
        } else {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
            [store requestAccessToEntityType:EKEntityTypeReminder completion:^(BOOL granted, NSError* error) {
                dispatch_semaphore_signal(sem);
            }];
#pragma clang diagnostic pop
        }
        dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);
    });
    return store;
}

// ISO 8601 date formatter with fractional seconds
static NSDateFormatter* get_iso_formatter(void) {
    static NSDateFormatter* fmt = nil;
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        fmt = [[NSDateFormatter alloc] init];
        [fmt setDateFormat:@"yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"];
        [fmt setTimeZone:[NSTimeZone timeZoneWithName:@"UTC"]];
        [fmt setLocale:[[NSLocale alloc] initWithLocaleIdentifier:@"en_US_POSIX"]];
    });
    return fmt;
}

// ISO 8601 date parser (handles both with and without fractional seconds)
static NSDate* parse_iso_date(const char* str) {
    if (!str) return nil;
    NSString* s = [NSString stringWithUTF8String:str];
    NSDate* d = [get_iso_formatter() dateFromString:s];
    if (d) return d;
    // Try without fractional seconds
    NSISO8601DateFormatter* iso = [[NSISO8601DateFormatter alloc] init];
    return [iso dateFromString:s];
}

static NSString* format_date(NSDate* date) {
    if (!date) return nil;
    return [get_iso_formatter() stringFromDate:date];
}

// Synchronous fetch of all reminders for given calendars (nil = all)
static NSArray<EKReminder*>* fetch_all_reminders(NSArray<EKCalendar*>* calendars) {
    EKEventStore* store = get_store();
    NSPredicate* predicate = [store predicateForRemindersInCalendars:calendars];
    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    __block NSArray<EKReminder*>* result = @[];
    [store fetchRemindersMatchingPredicate:predicate completion:^(NSArray<EKReminder*>* reminders) {
        result = reminders ?: @[];
        dispatch_semaphore_signal(sem);
    }];
    dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);
    return result;
}

// Convert an EKReminder to an NSDictionary for JSON serialization
static NSDictionary* reminder_to_dict(EKReminder* r) {
    NSMutableDictionary* d = [NSMutableDictionary dictionary];

    d[@"id"] = [NSString stringWithFormat:@"x-apple-reminder://%@", r.calendarItemIdentifier];
    d[@"name"] = r.title ?: @"";
    d[@"listName"] = r.calendar.title ?: @"";
    d[@"completed"] = @(r.isCompleted);
    d[@"flagged"] = @NO; // EventKit doesn't expose flagged
    d[@"priority"] = @(r.priority);

    // Due date from date components
    if (r.dueDateComponents) {
        NSCalendar* cal = [NSCalendar currentCalendar];
        NSDate* dueDate = [cal dateFromComponents:r.dueDateComponents];
        if (dueDate) {
            d[@"dueDate"] = format_date(dueDate);
        }
    }

    // Remind me date from first alarm
    if (r.alarms.count > 0) {
        EKAlarm* alarm = r.alarms.firstObject;
        if (alarm.absoluteDate) {
            d[@"remindMeDate"] = format_date(alarm.absoluteDate);
        }
    }

    if (r.completionDate) {
        d[@"completionDate"] = format_date(r.completionDate);
    }
    if (r.creationDate) {
        d[@"creationDate"] = format_date(r.creationDate);
    }
    if (r.lastModifiedDate) {
        d[@"modDate"] = format_date(r.lastModifiedDate);
    }

    d[@"body"] = r.notes ?: [NSNull null];

    return d;
}

// Serialize an array of dictionaries to JSON C string
static char* to_json(id obj) {
    NSError* error = nil;
    NSData* data = [NSJSONSerialization dataWithJSONObject:obj options:0 error:&error];
    if (!data) {
        set_error([NSString stringWithFormat:@"JSON serialization failed: %@", error.localizedDescription]);
        return NULL;
    }
    NSString* str = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
    return strdup([str UTF8String]);
}

// --- Public API ---

char* ek_fetch_lists(void) {
    @autoreleasepool {
        set_error(nil);
        EKEventStore* store = get_store();
        NSArray<EKCalendar*>* calendars = [store calendarsForEntityType:EKEntityTypeReminder];

        // Fetch all reminders to count per-list
        NSArray<EKReminder*>* allReminders = fetch_all_reminders(nil);
        NSMutableDictionary<NSString*, NSNumber*>* counts = [NSMutableDictionary dictionary];
        for (EKReminder* r in allReminders) {
            NSString* calId = r.calendar.calendarIdentifier;
            counts[calId] = @([counts[calId] integerValue] + 1);
        }

        NSMutableArray* result = [NSMutableArray array];
        for (EKCalendar* cal in calendars) {
            [result addObject:@{
                @"id": cal.calendarIdentifier,
                @"name": cal.title,
                @"count": counts[cal.calendarIdentifier] ?: @0
            }];
        }

        return to_json(result);
    }
}

char* ek_fetch_reminders(const char* list_name,
                         const char* completed_filter,
                         const char* search_query,
                         const char* due_before,
                         const char* due_after) {
    @autoreleasepool {
        set_error(nil);
        EKEventStore* store = get_store();

        // Find calendar for list filter
        NSArray<EKCalendar*>* cals = nil;
        if (list_name) {
            NSString* ln = [[NSString stringWithUTF8String:list_name] lowercaseString];
            NSMutableArray<EKCalendar*>* matched = [NSMutableArray array];
            for (EKCalendar* cal in [store calendarsForEntityType:EKEntityTypeReminder]) {
                if ([[cal.title lowercaseString] isEqualToString:ln]) {
                    [matched addObject:cal];
                }
            }
            if (matched.count == 0) {
                set_error([NSString stringWithFormat:@"list not found: %s", list_name]);
                return NULL;
            }
            cals = matched;
        }

        NSArray<EKReminder*>* allReminders = fetch_all_reminders(cals);

        // Parse date filters
        NSDate* dueBeforeDate = due_before ? parse_iso_date(due_before) : nil;
        NSDate* dueAfterDate = due_after ? parse_iso_date(due_after) : nil;

        // Search query
        NSString* query = search_query ? [[NSString stringWithUTF8String:search_query] lowercaseString] : nil;

        NSMutableArray* result = [NSMutableArray array];
        for (EKReminder* r in allReminders) {
            // Completed filter
            if (completed_filter) {
                if (strcmp(completed_filter, "true") == 0 && !r.isCompleted) continue;
                if (strcmp(completed_filter, "false") == 0 && r.isCompleted) continue;
            }

            // Search filter
            if (query) {
                NSString* titleLower = [(r.title ?: @"") lowercaseString];
                NSString* notesLower = [(r.notes ?: @"") lowercaseString];
                if (![titleLower containsString:query] && ![notesLower containsString:query]) {
                    continue;
                }
            }

            // Due date range filter
            if (dueBeforeDate || dueAfterDate) {
                if (!r.dueDateComponents) continue;
                NSDate* dueDate = [[NSCalendar currentCalendar] dateFromComponents:r.dueDateComponents];
                if (!dueDate) continue;
                if (dueBeforeDate && [dueDate compare:dueBeforeDate] == NSOrderedDescending) continue;
                if (dueAfterDate && [dueDate compare:dueAfterDate] == NSOrderedAscending) continue;
            }

            [result addObject:reminder_to_dict(r)];
        }

        return to_json(result);
    }
}

char* ek_get_reminder(const char* target_id) {
    @autoreleasepool {
        set_error(nil);
        if (!target_id) {
            set_error(@"target ID is required");
            return NULL;
        }

        NSArray<EKReminder*>* allReminders = fetch_all_reminders(nil);
        NSString* target = [[NSString stringWithUTF8String:target_id] uppercaseString];

        for (EKReminder* r in allReminders) {
            NSString* uuid = [r.calendarItemIdentifier uppercaseString];
            NSString* fullId = [NSString stringWithFormat:@"x-apple-reminder://%@", r.calendarItemIdentifier];

            if ([[fullId uppercaseString] isEqualToString:target] ||
                [uuid isEqualToString:target] ||
                [uuid hasPrefix:target]) {
                return to_json(reminder_to_dict(r));
            }
        }

        set_error([NSString stringWithFormat:@"reminder not found: %s", target_id]);
        return NULL;
    }
}
