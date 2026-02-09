import EventKit
import Foundation

// MARK: - Models

struct ListInfo: Codable {
    let id: String
    let name: String
    let count: Int
}

struct ReminderInfo: Codable {
    let id: String
    let name: String
    let listName: String
    let completed: Bool
    let flagged: Bool
    let priority: Int
    let dueDate: String?
    let remindMeDate: String?
    let completionDate: String?
    let creationDate: String?
    let modDate: String?
    let body: String?
}

// MARK: - Helpers

let isoFormatter: ISO8601DateFormatter = {
    let f = ISO8601DateFormatter()
    f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
    return f
}()

func fmtDate(_ date: Date?) -> String? {
    guard let d = date else { return nil }
    return isoFormatter.string(from: d)
}

func toReminderInfo(_ r: EKReminder) -> ReminderInfo {
    var dueDate: String? = nil
    if let comps = r.dueDateComponents, let d = Calendar.current.date(from: comps) {
        dueDate = fmtDate(d)
    }
    var remindDate: String? = nil
    if let alarms = r.alarms, let alarm = alarms.first, let d = alarm.absoluteDate {
        remindDate = fmtDate(d)
    }
    // EventKit doesn't expose flagged status directly.
    // We use priority as a proxy: in the Reminders app, flagged items have priority > 0
    // but this is not always accurate. The Go layer will overlay JXA flagged data when needed.
    return ReminderInfo(
        id: "x-apple-reminder://\(r.calendarItemIdentifier)",
        name: r.title ?? "",
        listName: r.calendar.title,
        completed: r.isCompleted,
        flagged: false,
        priority: r.priority,
        dueDate: dueDate,
        remindMeDate: remindDate,
        completionDate: fmtDate(r.completionDate),
        creationDate: fmtDate(r.creationDate),
        modDate: fmtDate(r.lastModifiedDate),
        body: r.notes
    )
}

// MARK: - Store Access

let store = EKEventStore()

func fetchAllReminders(calendars: [EKCalendar]?) -> [EKReminder] {
    let predicate = store.predicateForReminders(in: calendars)
    let semaphore = DispatchSemaphore(value: 0)
    var result: [EKReminder] = []
    store.fetchReminders(matching: predicate) { reminders in
        result = reminders ?? []
        semaphore.signal()
    }
    semaphore.wait()
    return result
}

// MARK: - Commands

func cmdLists() {
    let calendars = store.calendars(for: .reminder)
    let allReminders = fetchAllReminders(calendars: nil)

    var counts: [String: Int] = [:]
    for r in allReminders {
        counts[r.calendar.calendarIdentifier, default: 0] += 1
    }

    var lists: [ListInfo] = []
    for cal in calendars {
        lists.append(ListInfo(
            id: cal.calendarIdentifier,
            name: cal.title,
            count: counts[cal.calendarIdentifier] ?? 0
        ))
    }

    let encoder = JSONEncoder()
    if let data = try? encoder.encode(lists) {
        print(String(data: data, encoding: .utf8)!)
    }
}

func cmdReminders(listName: String?, completedFilter: String?,
                   searchQuery: String?, dueBefore: String?, dueAfter: String?) {
    var cals: [EKCalendar]? = nil
    if let ln = listName {
        let allCals = store.calendars(for: .reminder)
        let matched = allCals.filter { $0.title.lowercased() == ln.lowercased() }
        if matched.isEmpty {
            fputs("{\"error\":\"list not found: \(ln)\"}\n", stderr)
            exit(1)
        }
        cals = matched
    }

    let allReminders = fetchAllReminders(calendars: cals)

    var dueBeforeDate: Date? = nil
    var dueAfterDate: Date? = nil
    if let db = dueBefore {
        dueBeforeDate = isoFormatter.date(from: db) ?? ISO8601DateFormatter().date(from: db)
    }
    if let da = dueAfter {
        dueAfterDate = isoFormatter.date(from: da) ?? ISO8601DateFormatter().date(from: da)
    }

    let query = searchQuery?.lowercased()

    var result: [ReminderInfo] = []
    for r in allReminders {
        if let cf = completedFilter {
            if cf == "true" && !r.isCompleted { continue }
            if cf == "false" && r.isCompleted { continue }
        }

        if let q = query {
            let nameMatch = (r.title ?? "").lowercased().contains(q)
            let bodyMatch = (r.notes ?? "").lowercased().contains(q)
            if !nameMatch && !bodyMatch { continue }
        }

        if dueBeforeDate != nil || dueAfterDate != nil {
            guard let comps = r.dueDateComponents,
                  let dueDate = Calendar.current.date(from: comps) else {
                continue
            }
            if let before = dueBeforeDate, dueDate > before { continue }
            if let after = dueAfterDate, dueDate < after { continue }
        }

        result.append(toReminderInfo(r))
    }

    let encoder = JSONEncoder()
    if let data = try? encoder.encode(result) {
        print(String(data: data, encoding: .utf8)!)
    }
}

func cmdGet(targetId: String) {
    let allReminders = fetchAllReminders(calendars: nil)
    let upper = targetId.uppercased()

    for r in allReminders {
        let uuid = r.calendarItemIdentifier.uppercased()
        let fullId = "x-apple-reminder://\(r.calendarItemIdentifier)"
        if fullId == targetId || uuid == upper || uuid.hasPrefix(upper) {
            let info = toReminderInfo(r)
            let encoder = JSONEncoder()
            if let data = try? encoder.encode(info) {
                print(String(data: data, encoding: .utf8)!)
            }
            return
        }
    }

    fputs("{\"error\":\"reminder not found: \(targetId)\"}\n", stderr)
    exit(1)
}

// MARK: - Main

let args = CommandLine.arguments

if args.count < 2 {
    fputs("Usage: reminders-helper <command> [options]\n", stderr)
    fputs("Commands: lists, reminders, get\n", stderr)
    exit(1)
}

switch args[1] {
case "lists":
    cmdLists()

case "reminders":
    var listName: String? = nil
    var completed: String? = nil
    var search: String? = nil
    var dueBefore: String? = nil
    var dueAfter: String? = nil

    var i = 2
    while i < args.count {
        switch args[i] {
        case "--list":
            i += 1
            if i < args.count { listName = args[i] }
        case "--completed":
            i += 1
            if i < args.count { completed = args[i] }
        case "--search":
            i += 1
            if i < args.count { search = args[i] }
        case "--due-before":
            i += 1
            if i < args.count { dueBefore = args[i] }
        case "--due-after":
            i += 1
            if i < args.count { dueAfter = args[i] }
        default:
            break
        }
        i += 1
    }

    cmdReminders(listName: listName, completedFilter: completed,
                  searchQuery: search, dueBefore: dueBefore, dueAfter: dueAfter)

case "get":
    if args.count < 3 {
        fputs("Usage: reminders-helper get <id>\n", stderr)
        exit(1)
    }
    cmdGet(targetId: args[2])

default:
    fputs("Unknown command: \(args[1])\n", stderr)
    exit(1)
}
