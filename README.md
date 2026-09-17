# طهوریان — وب‌سایت رسمی دکتر حسین طهوریان

وب‌سایت شرکتی/شخصی دکتر حسین طهوریان (مشاور و راهبر کسب‌وکار، هلدینگ و استارتاپ) با
استک **Go (زبان v/chi)** + **SQLite** در بک‌اند و **HTML/Template + Tailwind +
Alpine.js + htmx + Lucide** در فرانت‌اند.

---

## ✨ امکانات

- **صفحات عمومی**: خانه، درباره، خدمات، پروژه‌ها (گالری)، مقالات + صفحه جزئیات مقاله، تیم، تماس
- **فرم تماس با htmx**: ارسال غیرهمزمان + توست (نوتیفیکیشن) + محدودسازی نرخ (Rate Limit)
- **اعلان تلگرام**: تمام اطلاعات فرم تماس به‌صورت خودکار به ایدی/گروه تلگرام شما ارسال می‌شود
- **پنل مدیریت** تحت نشانی `/admin`:
  - ورود/خروج با سشن امن و Cookie امن (HttpOnly + SameSite)
  - حفاظت CSRF بر درخواست‌های تغییردهنده
  - داشبورد (آمار مقالات)، مدیریت مقالات (افزودن/ویرایش/حذف)
- **بک‌اند داده**: SQLite پیش‌فرض (انباشت مداوم)، رفع جداول با Migration + Seed خودکار داده اولیه
- **بهینه موبایل**: طراحی دوباره‌پذیر با Breakpointهای موبایل، هدف‌های لمسی بزرگ،
  غیرفعال‌سازی افکت‌های Hover روی المان‌های لمسی
- **واب (Offline) زیرساخت**: کتابخانه‌های JS (Alpine, htmx, Lucide) به‌صورت محلی (local) سرو می‌شوند

---

## 🧰 تکنولوژی‌ها

| لایه | تکنولوژی |
|---|---|
| زبان | [Go (v / chi)](https://go-chi.dev) — runtime 1.27+ |
| روتینگ/وب | chi v5 |
| قالب | `html/template` |
| دیتابیس | SQLite (درایور glebarez/go-sqlite) + Migration جدا |
| سشن | پیاده‌سازی اختصاصی درون‌حافظه‌ای با پاکسازی خودکار |
| فرانت | Tailwind CSS + Alpine.js + htmx + Lucide (آیکن) |

---

## ⚙ راه‌اندازی

### پیش‌نیاز
- نصب Go runtime (نسخه ۱٫۲۵ به بالا)

### گام‌ها

```bash
# نصب وابستگی‌ها (بار اول)
go mod sync

# ساخت و اجرا
go run cmd/server/main.go
```

پس از اجرا مرورگر را در `http://localhost:8080` باز کنید.

### دست (Makefile)
```bash
make run        # اجرا سرور
make build      # ساخت (خروجی: bin/server)
make test       # اجرا تمامی تست‌ها
make migrate-up # اعمال Migration‌های goose
```

---

## 🔐 متغیرهای محیطی (`.env`)

فایل `.env.example` را کپی کرده و مقادیر را تنظیم کنید:

```dotenv
PORT=8080
ENV=development

# "sqlite" = دیتابیس دائمی (پیشنهادشده) | "memory" = حافظه موقتی
DB_TYPE=sqlite
DB_DSN=tahoorian.db

# کلید سشن — در پروداکشن یک مقدار رندم و بلند قرار دهید!
SESSION_SECRET=change-me-in-production

# حساب ورود به پنل مدیریت
ADMIN_USER=admin
ADMIN_PASS=admin

# اعلان تلگرام برای فرم تماس (اختیاری — اگر خالی باشد ارسال نمی‌شود)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

> ⚠ در استقرار صحیح `SESSION_SECRET`، `ADMIN_USER` و `ADMIN_PASS` را تغییر دهید.
> مقادیر پیش‌فرض فقط برای محیط توسعه است.

---

## 🗄 دیتابیس و Migration

- فایل‌های Migration در پوشه `migrations/` (سینتکس goose) قرار دارد.
- در اجرا اول، سرور به‌صورت خودکار:
  1. جداول را می‌سازد (معادل `001_initial_schema.sql`)
  2. داده اولیه را Seed می‌کند (بخش `about`، خدمات، آمار، اطلاعات تماس، گالری، مقالات)
- Seed **idempotent** است: اگر جدول خالی نباشد، دوباره داده اضافه نمی‌شود؛ پس هر اجرا امن است.
- دیتابیس پیش‌فرض `tahoorian.db` است (با `DB_DSN` قابل تغییر).

### اعمال Migration به‌صورت دست
```bash
make migrate-up
make migrate-down
```

---

## 🧪 تست

```bash
go test ./...
```

شامل: پروژه handler (روت‌های عمومی + جریان ورود/CSRF)، سرویس، و repository
(SQLite CRUD + خودکار Seed و idempotent بودن آن).

---

## 📁 ساختار پروژه

```
cmd/server/main.go          # نقطه ورود، روتینگ، انتخاب Repository
internal/config/            # خواندن تنظیمات از .env / محیط
internal/domain/            # مدل‌های دامنه (Article, Service, Contact, ...)
internal/handler/           # هندلر صفحات عمومی + پنل مدیریت
internal/service/           # لایه منطق کسب‌وکار
internal/notification/       # اعلان تلگرام برای فرم تماس
internal/repository/        # انتفیس Repository + پیاده‌سازی memory / sqlite
internal/repository/sqlite/ # schema migration, queries, seed
internal/middleware/        # logger, recoverer, session
migrations/                 # اسکریپت‌های migration (goose)
web/templates/              # قالب‌های HTML (layout, pages, partials, admin)
web/static/                 # CSS, JS, vendor (محلی), images
web/static/vendor/js/       # Alpine/htmx/Lucide سرو محلی
Makefile                    # شورت‌کات ساخت و تست
```

---

## 🔐 امنیت

- سشن‌ها با کلید امن `SESSION_SECRET` تنظیم می‌شوند (Cookie با `HttpOnly` و `SameSite=Lax`).
- رمز عبور ادمین از `ADMIN_USER`/`ADMIN_PASS` و فقط از environ خوانده می‌شود.
- حفاظت CSRF روی همه فرم‌های تغییردهنده پنل مدیریت.
- فرم تماس دارای محدودسازی نرخ (Rate-Limit) بر اساس IP.

---

## 📩 اعلان تلگرام برای فرم تماس

وقتی مشتری فرم تماس را پر و موفق ارسال کند، تمام اطلاعات او (نام، تلفن، ایمیل،
موضوع و متن پیام + زمان) در پس‌زمینه به یک چت تلگرام ازطریق ربات ارسال می‌شود.

### راه‌اندازی (یک بار)
1. در تلگرام با **@BotFather** چت کنید و یک ربات بسازید تا **توکن** دریافت کنید.
2. برای پیدا کردن **Chat ID**: به ربات یک پیام بفرستید (یا ربات را به گروه هدف اضافه کنید)
   و سپس از ربات‌های مثل **@userinfobot** یا `getUpdates` ایدی را بخوانید.
3. مقادیر را در `.env` قرار دهید:
   ```
   TELEGRAM_BOT_TOKEN=123456:ABC-xxxx
   TELEGRAM_CHAT_ID=-100123456789
   ```
4. سرور را اجرا کنید. ارسال رابطه با موفقیت در لاگ `contact submitted to telegram` ظاهر می‌شود.

> اگر `TELEGRAM_BOT_TOKEN` یا `TELEGRAM_CHAT_ID` خالی باشد، کل قابلیت به‌صورت امن غیرفعال
> می‌ماند و فرم همانند قبلی کار می‌کند (هیچ خطا به کاربر نموده نشود).

---

## 🚀 استقرار (Deployment)

1. ساخت بیناری: `make build`
2. تنظیم `.env` با مقادیر امن (ر.ک فوق).
3. اجرا با `bin/server` (یا مدیریت پروسه با systemd/supervisor).
4. برای اجرا دیوارند نهادی Reverse-Proxy (مثل Caddy/Nginx) با HTTPS پیشنهاد می‌شود.
5. سرور در `ENV != development` کش‌پذیری (Cache-Control) را فعال می‌کند.

> نکته Tailwind: در نسخه فعلی Tailwind از CDN لود می‌شود (موتور JIT کلاس‌ها را runtime
> می‌سازد). برای استقرار کاملاً آفلاین، Tailwind را پیش‌ساخت با ابزار CLI و لینک محلی کنید
> (فایل `web/static/css/tailwind.css` نمونه استقرار از پیش موجود دارد).

---

## 📄 مجوز

پروژه داخلی می‌باشد؛ استفاده تجاری با هماهنگی مالک وب‌سایت انجام شود.