import express, { Request, Response, NextFunction } from 'express';
import cookieParser from 'cookie-parser';
import path from 'path';
import crypto from 'crypto';
import dotenv from 'dotenv';

dotenv.config();

const app = express();
const PORT = parseInt(process.env.PORT || '3000', 10);
const HOST = '0.0.0.0';
const ADMIN_USER = process.env.ADMIN_USER || 'admin';
const ADMIN_PASS = process.env.ADMIN_PASS || '09152491524';
const SESSION_SECRET = process.env.SESSION_SECRET || 'tahoorian-secret-2026';

app.set('view engine', 'ejs');
app.set('views', path.join(process.cwd(), 'views'));
app.set('trust proxy', 1);

// Security headers middleware
app.use((req, res, next) => {
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'SAMEORIGIN');
  res.setHeader('X-XSS-Protection', '1; mode=block');
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
  next();
});

app.use(express.urlencoded({ extended: true }));
app.use(express.json());
app.use(cookieParser(SESSION_SECRET));

// Serve static assets from /web/static
app.use('/static', express.static(path.join(process.cwd(), 'web', 'static'), {
  maxAge: process.env.ENV === 'development' ? 0 : '1d'
}));

// Types & Models
interface Article {
  id: number;
  title: string;
  slug: string;
  category: string;
  summary: string;
  content: string;
  image_url: string;
  author: string;
  is_published: boolean;
  keywords?: string;
  reading_time?: string;
  created_at: Date;
  updated_at: Date;
  formattedDate?: string;
  formattedDateShort?: string;
}

interface ArticleCategory {
  id: number;
  slug: string;
  title: string;
  created_at: Date;
  formattedDateShort?: string;
}

interface Service {
  id: number;
  title: string;
  slug: string;
  description: string;
  content: string;
  image_url: string;
  is_active: boolean;
}

interface Stat {
  id: number;
  label: string;
  value: number;
  suffix: string;
}

interface GalleryItem {
  id: number;
  title: string;
  image_url: string;
  category: string;
  sort_order: number;
}

interface Contact {
  id: number;
  name: string;
  email: string;
  phone: string;
  subject: string;
  message: string;
  created_at: Date;
}

// In-Memory Database Store (Seeded mirroring Go SQLite seed)
let nextArticleId = 4;
let nextCategoryId = 5;
let nextContactId = 1;

const db = {
  about: {
    title: 'درباره ما',
    content: 'دکتر حسین طهوریان با بیش از ۲۰ سال تجربه در مدیریت و راهبری کسب‌وکارها، همواره در مسیر تحول و توسعه کشور گام برداشته است.',
    image_url: '/static/images/magnific_background-remover_cpvKmif0eP.png',
  },
  contact_info: {
    address: 'خیابان ولیعصر، برج تجاری آرین، طبقه ۱۲',
    phone: '۰۲۱-۲۲۰۱۴۵۶۷',
    email: 'info@tahoorian.ir',
    work_hours: 'شنبه تا چهارشنبه ۹:۰۰ الی ۱۷:۰۰',
  },
  services: <Service[]>[
    {
      id: 1,
      title: 'مشاوره مدیریت',
      slug: 'management-consulting',
      description: 'مشاوره تخصصی در زمینه مدیریت کسب‌وکار و بهبود عملکرد',
      content: '',
      image_url: '',
      is_active: true,
    },
    {
      id: 2,
      title: 'توسعه کسب‌وکار',
      slug: 'business-development',
      description: 'راهبری توسعه کسب‌وکار و ورود به بازارهای جدید',
      content: '',
      image_url: '',
      is_active: true,
    },
  ],
  stats: <Stat[]>[
    { id: 1, label: 'سال تجربه', value: 20, suffix: '+' },
    { id: 2, label: 'پروژه موفق', value: 150, suffix: '+' },
    { id: 3, label: 'مشتری راضی', value: 85, suffix: '+' },
    { id: 4, label: 'تیم متخصص', value: 45, suffix: '+' },
  ],
  gallery: <GalleryItem[]>[
    {
      id: 1,
      title: 'هلدینگ سرآمد سرمایه ایلیا',
      image_url: '/static/images/37-100x100.png',
      category: 'هلدینگ',
      sort_order: 1,
    },
    {
      id: 2,
      title: 'خدمات مشاوره‌ای معاد هلدینگ',
      image_url: '/static/images/32-100x100.png',
      category: 'خدمات حقوقی',
      sort_order: 2,
    },
    {
      id: 3,
      title: 'گروه مشاوره علمی پتنت',
      image_url: '/static/images/34-100x100.png',
      category: 'دانش‌بنیان',
      sort_order: 3,
    },
  ],
  categories: <ArticleCategory[]>[
    { id: 1, slug: 'traditional-medicine', title: 'طب سنتی و رگ‌گیری', created_at: new Date('2024-01-01') },
    { id: 2, slug: 'innovation', title: 'ثبت اختراع و نوآوری', created_at: new Date('2024-01-02') },
    { id: 3, slug: 'investment', title: 'سرمایه‌گذاری و هلدینگ', created_at: new Date('2024-01-03') },
    { id: 4, slug: 'consulting', title: 'مشاوره تخصصی مدیریت', created_at: new Date('2024-01-04') },
    { id: 5, slug: 'strategy', title: 'استراتژی کلان', created_at: new Date('2024-01-05') },
    { id: 6, slug: 'leadership', title: 'رهبری سازمان', created_at: new Date('2024-01-06') },
  ],
  articles: <Article[]>[
    {
      id: 1,
      title: 'آموزش جامع صفر تا صد ثبت اختراع؛ راهنمای گام‌به‌گام از ایده تا اخذ گواهینامه رسمی و ثبت بین‌المللی PCT',
      slug: 'amoozesh-sabt-ekhtera-rahnamaye-jame',
      category: 'innovation',
      summary: 'راهنمای عملی و گام‌به‌گام ثبت اختراع در سامانه اداره مالکیت معنوی ایران و معاهده بین‌المللی PCT، آموزش نحوه نگارش اصولی ادعانامه (Claims)، استعلام پیشینه FTO و مدل‌های تجاری‌سازی صنعتی با هدایت دکتر حسین طهوریان.',
      image_url: '/static/images/اختراعات-دکتر-سایت.webp',
      author: 'دکتر حسین طهوریان',
      is_published: true,
      reading_time: '۱۲ دقیقه',
      keywords: 'آموزش ثبت اختراع, مراحل ثبت اختراع, شرایط ثبت اختراع در ایران, نحوه ثبت اختراع در سامانه مالکیت معنوی, نگارش ادعانامه اختراع, جستجوی پیشینه اختراع, ثبت اختراع بین المللی PCT, هزینه ثبت اختراع, تجاری سازی اختراع, گام ابتکاری و کاربرد صنعتی, دکتر حسین طهوریان',
      created_at: new Date('2024-05-01'),
      updated_at: new Date('2024-05-01'),
      content: `<div class="space-y-8 text-slate-800 leading-loose">
<p class="text-lg leading-relaxed text-slate-700">
ثبت اختراع (Patent Registration) مطمئن‌ترین و حقوقی‌ترین ابزار برای تبدیل یک ایده نوآورانه به یک دارایی ارزشمند و انحصاری تجاری است. در دنیای رقابتی امروز، هر ایده فنی تا پیش از آنکه به‌طور قانونی به ثبت برسد، به‌شدت در معرض سرقت، کپی‌برداری و از دست رفتن حق تقدم قرار دارد. اما <strong>چگونه اختراع خود را ثبت کنیم؟</strong> مراحل ثبت در اداره مالکیت معنوی چیست و چگونه می‌توان از اشتباهات مرگباری که منجر به رد شدن اظهارنامه می‌شوند جلوگیری کرد؟ در این راهنمای جامع، تجربیات عملی حاصل از ثبت ۵ اختراع ملی، صنعتی و مدال‌های جهانی را در قالب یک نقشه راه گام‌به‌گام بررسی می‌کنیم.
</p>

<div class="p-6 rounded-2xl bg-amber-50 border-r-4 border-amber-500 text-slate-800 my-6">
<h3 class="text-base font-black text-amber-950 mb-2">📌 پاسخ سریع و خلاصه به سوال «اختراع چیست و چه شرایطی دارد؟»</h3>
<p class="text-sm leading-relaxed">
اختراع حاصل فکر فرد یا افرادی است که برای نخستین‌بار راه‌حل فناورانه جدیدی را برای حل یک مشکل فنی در صنعت، کشاورزی یا پزشکی ارائه می‌دهد. برای ثبت قانونی، اختراع باید هم‌زمان واجد ۳ شرط اساسی باشد: <strong>۱. تازگی و نوآوری در سطح جهانی (Novelty)</strong>، <strong>۲. دارا بودن گام ابتکاری (Inventive Step)</strong> که برای متخصص آن رشته بدیهی نباشد، و <strong>۳. قابلیت کاربرد صنعتی (Industrial Applicability)</strong>.
</p>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۱. سه شرط حیاتی و قانونی برای ثبت اختراع</h2>
<p>
مطابق با قوانین سازمان جهانی مالکیت فکری (WIPO) و قانون ثبت اختراعات ایران، ادعای اختراع در صورتی پذیرفته می‌شود که معیارهای سه‌گانه زیر را اثبات نماید:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>تازگی در سطح جهان (Absolute Novelty):</strong> طرح شما نباید در هیچ کجای دنیا، در قالب کتاب، مقاله دانشگاهی، سخنرانی، نمایشگاه یا اینترنت تا قبل از تاریخ ثبت اظهارنامه افشا شده باشد. حتی افشای ایده توسط خود مخترع پیش از تسلیم اظهارنامه ممکن است به رد اختراع بیانجامد.</li>
<li><strong>گام ابتکاری (Non-Obviousness):</strong> راه‌حل ارائه‌شده نباید برای فردی با مهارت عادی در آن صنعت (Person Skilled in the Art) واضح و بدیهی باشد؛ بلکه باید حاوی یک جهش خلاقانه تکنولوژیک باشد.</li>
<li><strong>کاربرد صنعتی (Industrial Utility):</strong> طرح باید در عمل قابل ساخت یا استفاده در یکی از شاخه‌های صنعت، کشاورزی یا خدمات باشد و صرفاً در حد یک تئوری انتزاعی علمی باقی نماند.</li>
</ul>

<div class="p-5 rounded-2xl bg-rose-50 border border-rose-200 text-rose-950 text-sm my-6">
<strong>⚠️ چه مواردی اختراع محسوب نمی‌شوند؟</strong><br>
کشف‌های علمی، نظریه‌های ریاضی، روش‌های صرفاً تجاری یا حسابداری، نرم‌افزارهای کامپیوتری فاقد اثر سخت‌افزاری ملموس، و روش‌های درمان یا جراحی بدن انسان و حیوان (اگرچه ابزارها و داروهای به کار رفته در درمان قابل ثبت هستند).
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۲. گام اول: جستجوی پیشینه اختراع و استعلام آزادی عمل (FTO Search)</h2>
<p>
بزرگ‌ترین اشتباه مخترعان تازه‌کار، اقدام به ثبت اظهارنامه بدون جستجوی پیشینه جهانی است. بیش از ۷۰ درصد اظهارنامه‌هایی که رد می‌شوند، پیش‌تر در کشورهای دیگر به ثبت رسیده‌اند. پیش از هر اقدامی باید در پایگاه‌های داده معتبر بین‌المللی جستجو کنید:
</p>
<ol class="list-decimal pr-6 space-y-2 text-slate-700">
<li><strong>Google Patents:</strong> جامع‌ترین و سریع‌ترین موتور جستجوی پتنت در میان بیش از ۱۰۰ کشور جهان.</li>
<li><strong>WIPO Patentscope:</strong> پایگاه داده رسمی سازمان جهانی مالکیت فکری برای بررسی پتنت‌های بین‌المللی PCT.</li>
<li><strong>Espacenet:</strong> پایگاه داده اداره ثبت اختراعات اروپا با بیش از ۱۴۰ میلیون سند نوآوری.</li>
<li><strong>سامانه مالکیت معنوی ایران:</strong> بخش جستجوی اختراعات ثبت‌شده در داخل کشور.</li>
</ol>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۳. گام دوم: تدوین فنی مستندات چهارگانه اختراع</h2>
<p>
سند اختراع یک مدرک فنی-حقوقی بسیار حساس است. نقص در نگارش این اوراق باعث می‌شود کارشناسان اداره ثبت بلافاصله نقص مدارک اعلام کرده یا چتر حفاظتی اختراع شما به‌شدت آسیب‌پذیر شود:
</p>

<div class="grid grid-cols-1 md:grid-cols-2 gap-4 my-4">
<div class="p-4 rounded-xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 mb-1">۱. توصیف اختراع (Description)</h4>
<p class="text-xs text-slate-600 leading-relaxed">شرح کامل حوزه فنی، پیشینه و مشکلات راه‌حل‌های قبلی، شرح نحوه عملکرد و ذکر بهترین روش پیاده‌سازی به گونه‌ای که یک متخصص بتواند آن را بازآفرینی کند.</p>
</div>
<div class="p-4 rounded-xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 mb-1">۲. ادعانامه (Claims) — مهم‌ترین بخش</h4>
<p class="text-xs text-slate-600 leading-relaxed">قلب تپنده اختراع! مرزهای انحصاری حقوقی شما را مشخص می‌کند. ادعانامه‌ها به دو دسته مستقل (Independent) و وابسته (Dependent) تقسیم می‌شوند.</p>
</div>
<div class="p-4 rounded-xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 mb-1">۳. نقشه‌های فنی (Drawings)</h4>
<p class="text-xs text-slate-600 leading-relaxed">ترسیم دقیق صنعتی، شماتیک‌ها یا فلوچارت‌های فنی با شماره‌گذاری استاندارد قطعات بر اساس ارجاعات داخل متن توصیف.</p>
</div>
<div class="p-4 rounded-xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 mb-1">۴. خلاصه اختراع (Abstract)</h4>
<p class="text-xs text-slate-600 leading-relaxed">چکیده مختصر فنی بین ۵۰ تا ۱۵۰ کلمه که هدف اصلی و ویژگی کلیدی راه‌حل نوآورانه را تبیین می‌نماید.</p>
</div>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۴. گام سوم: مراحل ثبت نام و ثبت اظهارنامه در سامانه مالکیت معنوی</h2>
<p>
تمامی مراحل ثبت اختراع در ایران به صورت الکترونیکی از طریق پورتال مرکز مالکیت معنوی قوه قضاییه (iripo.ssaa.ir) انجام می‌گیرد:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>احراز هویت ثنا و ورود به سامانه:</strong> ورود از طریق پنجره ملی خدمات دولت هوشمند یا کد ثنا.</li>
<li><strong>تکمیل مشخصات اظهارنامه:</strong> درج مشخصات کامل مخترع (فرد حقیقی) و مالک اختراع (حقیقی یا حقوقی مانند دانشگاه یا شرکت).</li>
<li><strong>بارگذاری فایل‌های Word و PDF:</strong> فایل‌های توصیف، ادعانامه، نقشه و خلاصه باید دقیقاً بر اساس فرمت سازمان بارگذاری شوند.</li>
<li><strong>پرداخت هزینه قانونی ثبت اظهارنامه و دریافت کد پیگیری:</strong> پس از ثبت، یک کد رهگیری و تاریخ ثبت رسمی (Priority Date) به شما تخصیص می‌یابد.</li>
</ul>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۵. گام چهارم: بررسی کارشناسی، داوری علمی و دفاع تخصصی</h2>
<p>
پس از بررسی شکلی، کارشناس پرونده را جهت بررسی ماهوی به یکی از مراجع علمی ذی‌صلاح (دانشگاه‌های مرجع کشور، سازمان پژوهش‌های علمی و صنعتی و...) ارسال می‌کند. در این مرحله، مخترع باید در جلسه دفاع علمی حاضر شده و با تسلط کامل بر ادبیات موضوع، نوآوری و گام ابتکاری خود را در برابر داوران اثبات نماید. پس از تایید مرجع استعلام، پرونده جهت روزنامه رسمی و صدور گواهی ارسال می‌گردد.
</p>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۶. ثبت بین‌المللی اختراع (معاهده PCT و سیستم پاریس)</h2>
<p>
حقوق حاصل از ثبت اختراع در ایران، محدود به قلمرو جغرافیایی ایران است. اگر قصد صادرات فناوری یا جذب سرمایه‌گذار خارجی دارید، حداکثر ظرف <strong>۱۲ ماه</strong> از تاریخ ثبت اولیه در ایران (طبق کنوانسیون پاریس) می‌توانید با تسلیم یک تقاضانامه بین‌المللی بر اساس معاهده همکاری ثبت اختراع (Patent Cooperation Treaty - PCT)، حق تقدم خود را در بیش از ۱۵۰ کشور جهان به مدت ۳۰ الی ۳۱ ماه محفوظ نگه دارید.
</p>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۷. تجاری‌سازی صنعتی و تبدیل پتنت به ثروت پایدار</h2>
<p>
اخذ گواهینامه ثبت اختراع پایان راه نیست، بلکه آغاز خلق ارزش است. سه راهبرد اصلی برای تجاری‌سازی وجود دارد:
</p>
<ol class="list-decimal pr-6 space-y-2 text-slate-700">
<li><strong>فروش کامل حق امتیاز (Assignment):</strong> انتقال مالکیت پتنت به یک شرکت صنعتی در ازای مبلغ نقد مقطوع.</li>
<li><strong>اعطای مجوز بهره‌برداری (Licensing):</strong> واگذاری حق تولید به کارخانجات در ازای دریافت درصد مشخصی از فروش سالیانه (Royalty).</li>
<li><strong>تاسیس شرکت دانش‌بنیان (Spin-off):</strong> تولید مستقیم محصول و بهره‌مندی از معافیت‌های مالیاتی و تسهیلات صندوق نوآوری و شکوفایی.</li>
</ol>

<div class="mt-8 p-6 rounded-2xl bg-slate-900 text-white">
<h3 class="text-lg font-black text-amber-400 mb-2">همراهی تخصصی با دکتر حسین طهوریان در فرآیند ثبت اختراع</h3>
<p class="text-sm text-slate-300 leading-relaxed mb-4">
دکتر طهوریان با تجربه ثبت ۵ اختراع ملی و تجاری‌سازی صنعتی، از ارزیابی اولیه ایده و استعلام پیشینه FTO گرفته تا نگارش فنی ادعانامه‌ها و حضور در جلسات دفاع علمی و مذاکره با سرمایه‌گذاران همراه شما خواهد بود.
</p>
<a href="/contact" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-amber-500 hover:bg-amber-400 text-slate-950 font-black text-xs transition">
<span>درخواست مشاوره اختصاصی ثبت اختراع</span>
<i data-lucide="arrow-left" class="w-4 h-4"></i>
</a>
</div>
</div>`,
    },
    {
      id: 2,
      title: 'تسکین فوری و درمان قطعی دردهای عضلانی با رگ‌گیری تخصصی؛ راهنمای جامع آزادسازی نقاط ماشه‌ای، گرفتگی گردن، شانه و سیاتیک',
      slug: 'taskin-dardhaye-azalani-roghgiri-dastvarzesh',
      category: 'traditional-medicine',
      summary: 'تحلیل بالینی و جامع مکانیزم تسکین سریع دردهای عضلانی، رفع اسپاسم‌های گردن، کتف، فیله‌های کمر و سیاتیک با متد کهن رگ‌گیری و درمان دستی طب سنتی ایرانی بدون وابستگی به مسکن‌های شیمیایی توسط دکتر حسین طهوریان.',
      image_url: '/static/images/download%20(1)-1.webp',
      author: 'دکتر حسین طهوریان',
      is_published: true,
      reading_time: '۱۵ دقیقه',
      keywords: 'تسکین دردهای عضلانی, درمان گرفتگی عضلات, رگ گیری, رگ گیری طب سنتی, درمان اسپاسم عضلانی گردن و شانه, درمان دردهای ستون فقرات, تریگر پوینت یا نقاط ماشه ای, رگ گیری سیاتیک, آزادسازی میوفاشیال, درمان دستی دردهای اسکلتی, علت گرفتگی مداوم عضلات, دکتر حسین طهوریان',
      created_at: new Date('2024-05-02'),
      updated_at: new Date('2024-05-02'),
      content: `<div class="space-y-8 text-slate-800 leading-loose">
<p class="text-lg leading-relaxed text-slate-700">
دردهای عضلانی-اسکلتی (Musculoskeletal Pain) از شایع‌ترین عارضه‌های فرساینده در زندگی پرفشار امروزی هستند. نشستن‌های طولانی‌مدت در پشت میز، استرس‌های مداوم شغلی، وضعیت‌های بدنی ناصحیح (Poor Posture) و رانندگی‌های طولانی باعث ایجاد گره‌های عضلانی سفت و دردناکی در عضلات گردن، کول، پشت و ناحیه لومبار (کمر) می‌شوند. اغلب افراد برای فرار از این دردها به مصرف مسکن‌های ضدالتهابی غیراستروئیدی (NSAIDs) نظیر ژلوفن، دیکلوفناک یا ناپروکسن روی می‌آورند؛ غافل از اینکه این داروها تنها پیام هشدار مغز را بی‌حس می‌کنند و ریشه مکانیکی انسداد بافت همچنان باقی می‌ماند. <strong>رگ‌گیری تخصصی در طب سنتی ایرانی</strong> راهکاری بالینی و بدون دارو برای درمان ریشه‌ای این گرفتگی‌هاست.
</p>

<div class="p-6 rounded-2xl bg-emerald-50 border-r-4 border-emerald-600 text-slate-800 my-6">
<h3 class="text-base font-black text-emerald-950 mb-2">🩺 رگ‌گیری چیست و چگونه دردهای عضلانی را تسکین می‌دهد؟</h3>
<p class="text-sm leading-relaxed">
رگ‌گیری یکی از اصیل‌ترین و موثرترین شاخه‌های درمان دستی (Manual Therapy) در طب سنتی ایرانی است. این فن با شناسایی دقیق گره‌های انقباضی و نقاط ماشه‌ای (Trigger Points) در مسیر فاسیا، عروق خونی و پایانه‌های عصبی، با اعمال فشارهای نقطه‌ای و امتدادی کنترل‌شده، گره‌های فیبری را باز کرده و جریان راکد خون و لنف را فوراً برقرار می‌سازد.
</p>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۱. نقاط ماشه‌ای (Trigger Points) و چرخه معیوب درد و اسپاسم</h2>
<p>
هنگامی که فیبرهای یک ماهیچه به علت فشار مکانیکی، ضربه یا استرس عصبی در حالت انقباض مداوم باقی می‌مانند، یک کانون بیش‌انگیخته (Hyperirritable Spot) موسوم به «نقطه ماشه‌ای» پدید می‌آید. در این نقطه:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>ایسکمی موضعی (کاهش خون‌رسانی):</strong> رگ‌های مویرگی فشرده شده و اکسیژن کافی به بافت عضله نمی‌رسد.</li>
<li><strong>تجمع متابولیت‌های اسیدی:</strong> اسید لاکتیک و مواد التهابی (مانند پروستاگلاندین‌ها و برادی‌کینین) در محل تجمع می‌یابند.</li>
<li><strong>چرخه بی‌پایان اسپاسم:</strong> با اسیدی شدن محیط، گیرنده‌های درد تحریک شده و مغز پیام انقباض بیشتر صادر می‌کند؛ در نتیجه عضله سفت‌تر و دردناک‌تر می‌شود.</li>
</ul>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۲. تفاوت اساسی رگ‌گیری تخصصی با ماساژ ریلکسی</h2>
<p>
بسیاری از مراجعین رگ‌گیری را با ماساژ اشتباه می‌گیرند. در حالی که ماساژ بیشتر بر تسکین سطحی و آرامش پوست و لایه‌های رویی تمرکز دارد، <strong>رگ‌گیری یک مداخله بالینی عمقی</strong> است:
</p>
<div class="overflow-x-auto my-4">
<table class="w-full text-right text-xs border border-slate-200 rounded-xl overflow-hidden">
<thead class="bg-slate-100 text-slate-900 font-black">
<tr>
<th class="p-3">معیار</th>
<th class="p-3">ماساژ معمولی</th>
<th class="p-3">رگ‌گیری تخصصی طب سنتی</th>
</tr>
</thead>
<tbody class="divide-y divide-slate-200">
<tr>
<td class="p-3 font-bold">عمق نفوذ</td>
<td class="p-3">پوست و عضلات سطحی</td>
<td class="p-3 font-semibold text-emerald-800">لایه‌های عمقی عضلانی، فاسیا و عروق پیرامونی</td>
</tr>
<tr>
<td class="p-3 font-bold">رویکرد درمانی</td>
<td class="p-3">ریلکسیشن عمومی و موقت</td>
<td class="p-3 font-semibold text-emerald-800">آزادسازی مکانیکی انسدادها و شکستن گره‌های انقباضی</td>
</tr>
<tr>
<td class="p-3 font-bold">پایداری نتیجه</td>
<td class="p-3">چند ساعت تا چند روز</td>
<td class="p-3 font-semibold text-emerald-800">تسکین ماندگار و ریشه‌ای با اصلاح خون‌رسانی بافتی</td>
</tr>
</tbody>
</table>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۳. پنج اثر بیولوژیک و درمانی شگفت‌انگیز رگ‌گیری</h2>
<ol class="list-decimal pr-6 space-y-3 text-slate-700">
<li><strong>خون‌رسانی مجدد اکسیداتیو (Hyperemia):</strong> پس از اعمال فشار مکانیکی بر گره و رهاسازی آن، موجی از خون تازه سرشار از اکسیژن و مواد مغذی به بافت محروم می‌رسد.</li>
<li><strong>آزادسازی فاسیای درهم‌تنیده (Myofascial Release):</strong> غلاف همبندی فاسیا که عضلات را در بر گرفته نرم و انعطاف‌پذیر می‌شود.</li>
<li><strong>تخلیه متابولیت‌های التهابی و سموم بافتی:</strong> تحریک کانال‌های لنفاوی موجب تخلیه اسیدهای انباشته و کاهش فوری ورم و سنگینی در اندام‌ها می‌گردد.</li>
<li><strong>تحریک ترشح هورمون‌های اندورفین:</strong> اثر ضددرد طبیعی بدن فعال شده و بدون هیچ‌گونه عوارض دارویی، احساس سبکی و تسکین حاصل می‌شود.</li>
<li><strong>تنظیم مجدد گیرنده‌های حس عمقی (Proprioception):</strong> تون عضلانی طبیعی بدن احیا شده و فشار نابرابر از روی ستون فقرات و دیسک‌ها برداشته می‌شود.</li>
</ol>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۴. دردهای شایعی که با رگ‌گیری تخصصی درمان می‌شوند</h2>

<div class="grid grid-cols-1 md:grid-cols-2 gap-4 my-4">
<div class="p-5 rounded-2xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 text-sm mb-2 flex items-center gap-1.5">
<i data-lucide="check-circle" class="w-4 h-4 text-emerald-600"></i>
اسپاسم گردن، کول و شانه منجمد
</h4>
<p class="text-xs text-slate-600 leading-relaxed">
درد شایع مدیران و کارمندان ناشی از نگاه مداوم به مانیتور و گوشی. رگ‌گیری عضلات ذوزنقه‌ای (Trapezius) و بالابرنده کتف، دامنه حرکتی گردن را در همان جلسه نخست بازمی‌گرداند.
</p>
</div>

<div class="p-5 rounded-2xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 text-sm mb-2 flex items-center gap-1.5">
<i data-lucide="check-circle" class="w-4 h-4 text-emerald-600"></i>
دردهای سیاتیک و گرفتگی فیله‌های کمر
</h4>
<p class="text-xs text-slate-600 leading-relaxed">
در بسیاری از موارد، درد سیاتیک ناشی از گرفتگی عضله پیریفورمیس در باسن است که عصب را به دام انداخته است (Piriformis Syndrome). رگ‌گیری این عضله، فشار را از روی عصب سیاتیک آزاد می‌کند.
</p>
</div>

<div class="p-5 rounded-2xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 text-sm mb-2 flex items-center gap-1.5">
<i data-lucide="check-circle" class="w-4 h-4 text-emerald-600"></i>
سردردهای تنشی و میگرنی پس‌سری
</h4>
<p class="text-xs text-slate-600 leading-relaxed">
گره‌های عضلانی در قاعده جمجمه و عضلات گردن، جریان خون شریانی به سر را مختل کرده و منشأ سردردهای کوبنده می‌شوند که با رگ‌گیری تخصصی به سرعت تسکین می‌یابند.
</p>
</div>

<div class="p-5 rounded-2xl bg-slate-50 border">
<h4 class="font-bold text-slate-900 text-sm mb-2 flex items-center gap-1.5">
<i data-lucide="check-circle" class="w-4 h-4 text-emerald-600"></i>
خستگی مفرط شغلی و سندرم بی‌قراری عضلات
</h4>
<p class="text-xs text-slate-600 leading-relaxed">
احساس کوفتگی مداوم و سنگینی در پاها و شانه‌ها که با خواب نیز برطرف نمی‌شود؛ ناشی از رکود جریان لنفاوی و اخلاط غلیظ در بافت عضلانی است.
</p>
</div>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۵. تدابیر تکمیلی طب سنتی برای ماندگاری اثر رگ‌گیری</h2>
<p>
برای تضمین عدم بازگشت دردهای عضلانی، رگ‌گیری با پروتکل‌های تغذیه‌ای و سبک زندگی طب سنتی تلفیق می‌گردد:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>روغن‌مالی موضعی با روغن‌های گرم:</strong> استفاده از روغن سیاه‌دانه، بابونه یا بادام تلخ به باز شدن منافذ و تسریع خون‌رسانی کمک می‌کند.</li>
<li><strong>پرهیز از مصرف سردی‌جات افراطی:</strong> مصرف ماست، دوغ و ترشیجات به خصوص در ساعات پایانی شب، موجب انقباض و سردی عضلات کمر و گردن می‌شود.</li>
<li><strong>نوشیدن دمنوش‌های ضداسپاسم:</strong> دم‌کرده بادرنجبویه، اسطوخودوس و دارچین اثر آرام‌بخش بر سیستم عصبی مرکزی دارد.</li>
<li><strong>رعایت ارگونومی صحیح کار:</strong> تنظیم زاویه نگاه به مانیتور و برخاستن و انجام کشش‌های سبک پس از هر ۴۵ دقیقه نشستن مداوم.</li>
</ul>

<div class="mt-8 p-6 rounded-2xl bg-emerald-900 text-white">
<h3 class="text-lg font-black text-amber-300 mb-2">جلسات حضوری رگ‌گیری با دکتر حسین طهوریان</h3>
<p class="text-sm text-emerald-100 leading-relaxed mb-4">
کلیه جلسات رگ‌گیری مستقیماً توسط شخص دکتر طهوریان در محیط اختصاصی و کاملاً بهداشتی کلینیک انجام می‌شود. پیش از شروع، معاینه ساختار اسکلتی-عضلانی و بررسی شرح‌حال بالینی صورت می‌گیرد تا متناسب‌ترین درمان دستی پیاده‌سازی شود.
</p>
<a href="/contact" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-amber-400 hover:bg-amber-300 text-slate-950 font-black text-xs transition">
<span>رزرو نوبت جلسه ارزیابی و رگ‌گیری</span>
<i data-lucide="arrow-left" class="w-4 h-4"></i>
</a>
</div>
</div>`,
    },
    {
      id: 3,
      title: 'مبانی و اصول بنیادین طب سنتی ایرانی؛ راهنمای جامع مزاج‌شناسی، تعادل اخلاط اربعه و ۶ اصل حیاتی سلامت (ستّه ضروریه)',
      slug: 'teb-sonati-irani-mabani-mezaj-shenasi',
      category: 'traditional-medicine',
      summary: 'آشنایی جامع با حکمت پزشکی کهن ایران، ارکان اربعه، مزاج‌شناسی بالینی، اخلاط چهارگانه (دم، صفرا، بلغم، سودا) و ستّه ضروریه (شش اصل حیاتی حفظ سلامت و طول عمر) به قلم پژوهشگر طب سنتی دکتر حسین طهوریان.',
      image_url: '/static/images/12.webp',
      author: 'دکتر حسین طهوریان',
      is_published: true,
      reading_time: '۱۴ دقیقه',
      keywords: 'طب سنتی, طب سنتی ایرانی, مزاج شناسی, تشخیص طبع و مزاج, اخلاط اربعه, مزاج گرم و خشک, مزاج سرد و تر, سته ضروریه طب سنتی, پاکسازی کبد و گوارش در طب سنتی, سبک زندگی سالم طب سنتی, پیشگیری از بیماری ها, تدابیر فصول, دکتر حسین طهوریان',
      created_at: new Date('2024-05-03'),
      updated_at: new Date('2024-05-03'),
      content: `<div class="space-y-8 text-slate-800 leading-loose">
<p class="text-lg leading-relaxed text-slate-700">
مکتب <strong>طب سنتی ایرانی (Persian Traditional Medicine)</strong> که ریشه در آثار حکمای نامداری چون ابن‌سینا (ابوعلی سینا)، رازی و جرجانی دارد، یکی از غنی‌ترین و منسجم‌ترین دستگاه‌های پزشکی کل‌نگر (Holistic Medicine) در تاریخ بشریت است. تفاوت بنیادین طب ایرانی با پزشکی رایج امروزی، تمرکز عمیق آن بر <strong>پیشگیری بر درمان (حفظ‌الصحه)</strong> و نگریستن به انسان به مثابه یک کل هماهنگ با کیهان و طبیعت است. در این مکتب، هیچ درمانی بدون شناخت طبع و مزاج انحصاری بیمار و تعادل اخلاط درونی تجویز نمی‌شود. در این مقاله تخصصی، مبانی حکمت پزشکی ایرانی و ابزارهای کاربردی آن برای دستیابی به سلامت پایدار و طول عمر باکیفیت را مرور می‌کنیم.
</p>

<div class="p-6 rounded-2xl bg-amber-50 border-r-4 border-amber-600 text-slate-800 my-6">
<h3 class="text-base font-black text-amber-950 mb-2">🌿 اصل بنیادین طب سنتی: «طبیعت مدبره بدن»</h3>
<p class="text-sm leading-relaxed">
در نگرش طب سنتی، بدن انسان دارای یک هوشمندی درونی ذاتی به نام «طبیعت مدبره» (Vis Medicatrix Naturae) است که پیوسته در تلاش است تا تعادل زیستی (Homeostasis) را حفظ و بیماری‌ها را مهار کند. وظیفه پزشک و درمانگر، مبارزه با علائم نیست؛ بلکه برطرف کردن موانع و همراهی با طبیعت بدن از طریق تنظیم مزاج، تغذیه و درمان‌های دستی اصیل است.
</p>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۱. ارکان اربعه و کیفیات چهارگانه هستی</h2>
<p>
حکمای طب سنتی معتقد بودند که جهان مادی و بدن انسان از چهار عنصر اولیه یا «ارکان اربعه» شکل گرفته‌اند که هر یک حامل دو کیفیت اصلی هستند:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>آتش:</strong> گرم و خشک (سبک‌ترین رکن، مایه تحرک و هضم و سوخت‌وساز)</li>
<li><strong>هوا:</strong> گرم و تر (بخشنده لطافت، تنفس و انعطاف‌پذیری)</li>
<li><strong>آب:</strong> سرد و تر (عامل سیالیت، رطوبت‌بخشی و نرمی بافت‌ها)</li>
<li><strong>خاک:</strong> سرد و خشک (سنگین‌ترین رکن، ضامن استحکام، ثبات و فرم‌گیری اندام‌ها)</li>
</ul>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۲. اخلاط اربعه؛ مایعات چهارگانه حیاتی در گردش خون</h2>
<p>
هنگامی که غذا از مسیر معده عبور کرده و وارد کبد می‌شود (هضم دوم)، به چهار مایع حیاتی یا «خلط» تبدیل می‌گردد که بقای ارگان‌های زیستی را تضمین می‌کنند:
</p>

<div class="grid grid-cols-1 md:grid-cols-2 gap-4 my-4">
<div class="p-5 rounded-2xl bg-rose-50 border border-rose-200">
<h4 class="font-bold text-rose-950 text-sm mb-1">۱. خلط دم (خون) — گرم و تر</h4>
<p class="text-xs text-rose-800 leading-relaxed">سرخ‌فام و معتدل‌ترین خلط. حامل اصلی اکسیژن، حرارت و مواد مغذی به تمام سلول‌ها. افراد دموی‌مزاج، بدنی قوی، خوش‌مشرب، پرانرژی و چهره‌ای گلگون دارند.</p>
</div>

<div class="p-5 rounded-2xl bg-amber-50 border border-amber-200">
<h4 class="font-bold text-amber-950 text-sm mb-1">۲. خلط صفرا — گرم و خشک</h4>
<p class="text-xs text-amber-800 leading-relaxed">زردرنگ و سبک‌ترین خلط. وظیفه آن باز کردن عروق، تسهیل هضم چربی‌ها و برانگیختن انرژی است. صفراوی‌مزاجان پرجنب‌وجوش، تیزهوش، شجاع و زودخشم هستند.</p>
</div>

<div class="p-5 rounded-2xl bg-sky-50 border border-sky-200">
<h4 class="font-bold text-sky-950 text-sm mb-1">۳. خلط بلغم — سرد و تر</h4>
<p class="text-xs text-sky-800 leading-relaxed">سفید و شفاف. تشکیل‌دهنده مایعات مفصلی، بزاق و مایع مغزی-نخاعی. بلغمی‌مزاجان صبور، آرام، دارای پوستی روشن و مستعد انباشت چربی و کندی متابولیسم هستند.</p>
</div>

<div class="p-5 rounded-2xl bg-slate-100 border border-slate-300">
<h4 class="font-bold text-slate-950 text-sm mb-1">۴. خلط سودا — سرد و خشک</h4>
<p class="text-xs text-slate-800 leading-relaxed">تیره‌ترین و سنگین‌ترین خلط. رسوب‌کننده در استخوان‌ها، طحال و بافت‌های سخت. سوداوی‌مزاجان افرادی دقیق، متفکر، منظم، آینده‌نگر و مستعد اضطراب و افکار وسواسی‌اند.</p>
</div>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۳. تشخیص مزاج؛ ۱۰ شاخص بالینی (اجناس عشره)</h2>
<p>
برای تعیین طبع پایه (مزاج جبلی)، حکیم بر اساس ۱۰ نشانه اصلی بدن وضعیت اخلاط را می‌سنجد:
</p>
<ol class="list-decimal pr-6 space-y-2 text-slate-700">
<li><strong>ملمس (حرارت و رطوبت پوست در لمس):</strong> گرمی یا سردی، زبری یا نرمی دست و اندام‌ها.</li>
<li><strong>رنگ پوست چهره:</strong> زردی (غلبه صفرا)، سرخی (دم)، رنگ‌پریدگی و سفیدی (بلغم)، کدورت و تیرگی (سودا).</li>
<li><strong>ساختار عضلانی و اسکلتی:</strong> درشتی استخوان‌ها و توده عضلانی در برابر جثه ظریف و لاغر.</li>
<li><strong>موها:</strong> سرعت رشد، ضخامت، پیچیدگی و رنگ مو.</li>
<li><strong>کیفیت خواب و بیداری:</strong> خواب عمیق و طولانی (تری و سردی) در برابر خواب سبک و کوتاه‌مدت (گرمی و خشکی).</li>
<li><strong>وضعیت نبض:</strong> سرعت، عمق، تواتر و استحکام امواج نبض شریانی.</li>
<li><strong>فضولات بدنی:</strong> رنگ و بوی ادرار، مدفوع و عرق بدن.</li>
<li><strong>رفتارها و افعال نفسانی:</strong> سرعت تصمیم‌گیری، شجاعت، آرامش یا درون‌گرایی.</li>
</ol>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۴. ستّه ضروریه؛ شش اصل حیاتی سلامت و راز طول عمر ابن‌سینا</h2>
<p>
شاهکار مکتب طب سنتی ایران، نظریه <strong>ستّه ضروریه (شش امر ناگزیر حیات)</strong> است. این اصول شش‌گانه، سبک زندگی انسان را تشکیل می‌دهند و رعایت آنها سلامتی را تضمین و بیماری‌ها را ریشه‌کن می‌کند:
</p>

<div class="space-y-4 my-6">
<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۱. هوا و تنفس پاکیزه (الهواء المحیط):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">کیفیت هوای تنفسی، اکسیژن‌گیری کامل ریوی و هماهنگی با اقلیم زندگی اولین شرط تولید روح حیوانی و شادابی بافت‌هاست.</p>
</div>

<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۲. ماکول و مشروب (خوردنی‌ها و نوشیدنی‌ها):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">غذا باید دوای انسان باشد. رعایت تناسب غذا با طبع، خوب جویدن، پرهیز از درهم‌خوری و نخوردن آب یخ همراه با غذا ضامن سلامت کبد و معده است.</p>
</div>

<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۳. حرکت و سکون (ورزش و ارگونومی):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">ورزش معتدل باعث افزایش حرارت غریزی، دفع فضولات از تعریق و تقویت قلب می‌شود. بی‌تحرکی عامل شماره یک غلبه بلغم و افت عملکرد اعضاست.</p>
</div>

<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۴. خواب و بیداری (النوم و الیقظه):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">بهترین زمان خواب عمیق از ساعت ۱۰ شب تا طلوع آفتاب است؛ زمانی که کبد و سیستم ایمنی مشغول بازسازی و پاکسازی سموم زیستی هستند.</p>
</div>

<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۵. استفراغ و احتباس (دفع سموم و نگه‌داری ضروریات):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">تخلیه منظم روده‌ها، تعریق مناسب و پاکسازی مسیرهای دفعی؛ هم‌زمان با جلوگیری از هدررفت رطوبت و انرژی حیاتی بدن.</p>
</div>

<div class="p-4 rounded-xl bg-slate-50 border-r-4 border-emerald-600">
<strong class="text-slate-900 block text-sm">۶. اعراض نفسانی (احوال روحی و روانی):</strong>
<p class="text-xs text-slate-600 mt-1 leading-relaxed">شادی، غم، خشم، استرس و آرامش مستقیماً حرارت غریزی و هورمون‌ها را دگرگون می‌کنند. آرامش روان بالاترین داروی سلامت کالبد است.</p>
</div>
</div>

<h2 class="text-2xl font-black text-slate-900 border-b pb-3 pt-4">۵. پاکسازی‌های فصلی و هماهنگی با دگرگونی‌های اقلیم</h2>
<p>
طبیعت با گردش فصول تغییر مزاج می‌دهد؛ بهار گرم و تر است، تابستان گرم و خشک، پاییز سرد و خشک و زمستان سرد و تر. انسان سالم کسی است که رژیم غذایی و پوشش خود را با تغییر فصول سازگار کند:
</p>
<ul class="list-disc pr-6 space-y-2 text-slate-700">
<li><strong>بهار:</strong> زمان پاکسازی خون و کبد، کاهش حجم غذاهای سنگین گوشتی و مصرف سکنجبین و کاسنی.</li>
<li><strong>تابستان:</strong> پرهیز از تندی‌ها و سرخ‌کردنی‌های صفرازا، مصرف خنکی‌های طبیعی مانند لیموترش و هندوانه.</li>
<li><strong>پاییز:</strong> محافظت از پوست در برابر خشکی سوداوی، روغن‌مالی بدن با روغن بنفشه و بادام شیرین.</li>
<li><strong>زمستان:</strong> افزایش مصرف گرمی‌جات طبیعی مانند خرما، دارچین و زنجبیل جهت تقویت هضم معدی.</li>
</ul>

<div class="mt-8 p-6 rounded-2xl bg-amber-950 text-white">
<h3 class="text-lg font-black text-amber-300 mb-2">شناسنامه مزاج‌شناسی و تنظیم رژیم اختصاصی با دکتر حسین طهوریان</h3>
<p class="text-sm text-slate-200 leading-relaxed mb-4">
در کلینیک تخصصی دکتر طهوریان، پس از معاینه بالینی دقیق، شناسنامه اختصاصی مزاج‌شناسی به همراه پروتکل اصلاح تغذیه، دمنوش‌های درمانی و جلسات رگ‌گیری دستی جهت بازگرداندن توازن زیستی ارائه می‌گردد.
</p>
<a href="/contact" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-amber-400 hover:bg-amber-300 text-slate-950 font-black text-xs transition">
<span>درخواست مشاوره و ارزیابی طبع</span>
<i data-lucide="arrow-left" class="w-4 h-4"></i>
</a>
</div>
</div>`,
    },
  ],
  packages: [
    {
      id: 'roghgiri-health',
      theme: 'emerald',
      title: 'سلامت، طب سنتی و رگ‌گیری تخصصی',
      latinTitle: 'Executive Health, Roghgiri & Traditional Medicine',
      tag: 'اصالت طب سنتی و درمان دستی',
      icon: 'heart-pulse',
      isFeatured: false,
      audience: 'مدیران ارشد، کارآفرینان و افرادی که با دردهای مزمن ستون فقرات، گرفتگی‌های گردن و کتف، استرس مداوم و خستگی شغلی مواجهند.',
      description: 'رویکردی کاملاً طبیعی، علمی و منطبق بر گنجینه طب سنتی ایرانی و فنون کهن رگ‌گیری. این پکیج با تمرکز بر آزادسازی گره‌های عضلانی عمیق (Myofascial Trigger Points)، تعدیل اخلاط و پاکسازی مسیرهای عروقی، انرژی و تمرکز ازدست‌رفته را به جسم و روان شما بازمی‌گرداند.',
      features: [
        'معاینه بالینی مزاج و ارزیابی عارضه‌یابانه ستون فقرات و نقاط ماشه‌ای درد',
        'جلسات رگ‌گیری تخصصی با دست توسط دکتر طهوریان بدون داروهای شیمیایی',
        'آزادسازی فوری اسپاسم‌های عمقی گردن، کتف، پشت، فیله‌های کمر و سیاتیک',
        'تسهیل خون‌رسانی مویرگی و تخلیه لنفاوی برای کاهش ورم و سموم بافتی',
        'تنظیم برنامه غذایی و دمنوش‌های درمانی متناسب با طبع انحصاری شما',
        'آموزش ارگونومی و حرکات اصلاحی روزانه در محیط کار برای پیشگیری همیشگی از عود',
      ],
      deliverables: [
        'شناسنامه تشخیصی طبع و ساختار اسکلتی-عضلانی',
        'پروتکل اختصاصی اصلاح سبک زندگی و رژیم مزاجی',
        'برنامه تمرینات کششی اصلاحی اختصاصی محیط کار',
      ],
      processSteps: [
        { num: '۰۱', title: 'ارزیابی طبع و اسکن دردهای فیزیکی' },
        { num: '۰۲', title: 'جلسات رگ‌گیری و آزادسازی گره‌های انسدادی' },
        { num: '۰۳', title: 'اصلاح خون‌رسانی و تدابیر پاکسازی زیستی' },
        { num: '۰۴', title: 'پایش و ارائه راهنمای حفظ سلامت پایدار' },
      ],
      sessions: '۴ الی ۶ جلسه حضوری در کلینیک تخصصی',
      duration: 'هر جلسه ۵۰ الی ۶۰ دقیقه',
      location: 'کلینیک تخصصی دکتر طهوریان',
      guarantee: 'محرمانگی کامل پزشکی و همراهی انفرادی بدون واسطه',
      actionText: 'رزرو وقت ارزیابی و رگ‌گیری',
    },
    {
      id: 'venture-investment',
      theme: 'gold',
      title: 'سرمایه‌گذاری، مقیاس‌پذیری و توسعه هلدینگ',
      latinTitle: 'Venture Scaling, Investment & Holding Architecture',
      tag: 'پیشنهاد اول مدیران عامل و هلدینگ‌ها',
      icon: 'crown',
      isFeatured: true,
      ribbonText: 'پیشنهاد ویژه و استراتژیک برای جهش نمایی',
      audience: 'مالکان شرکت‌ها، بنیان‌گذاران استارتاپ‌های مقیاس‌پذیر و مدیران هلدینگ‌های صنعتی و بازرگانی در پی جذب سرمایه و بازمهندسی ساختار.',
      description: 'بسته جامع و همه‌جانبه برای بازآفرینی معماری مالی، تجدید ساختار دارایی‌ها، جذب سرمایه‌گذاران استراتژیک و خلق هم‌افزایی پایدار در شرکت‌های هلدینگ. از مدل‌سازی درآمدی تا حضور مستقیم در میز مذاکره سرمایه‌گذاری.',
      features: [
        'ارزش‌گذاری موشکافانه کسب‌وکار و آماده‌سازی مستندات و مدل مالی جذب سرمایه‌گذار',
        'طراحی ساختار حقوقی سهام‌داری، مدیریت ریسک هلدینگ و ورود شرکای جدید',
        'تدوین نقشه استراتژیک ۳ ساله برای تسخیر بازارهای هدف و افزایش سهم بازار',
        'خلق هم‌افزایی ارزش میان شرکت‌های اقماری و بهینه‌سازی جریان نقدینگی',
        'همراهی مستقیم دکتر طهوریان در جلسات حساس با سرمایه‌گذاران و هیئت‌مدیره',
        'طراحی ساختار حاکمیت شرکتی (Corporate Governance) و مدیریت ریسک سبد سرمایه',
      ],
      deliverables: [
        'کتابچه جامع ارزش‌گذاری رسمی و مستندات Investment Pitch',
        'مدل مالی داینامیک ۵ ساله و سناریوهای پیش‌بینی جریان نقدینگی',
        'اساسنامه و چارچوب حقوقی سهام‌داری و آیین‌نامه‌های حاکمیت شرکتی',
      ],
      processSteps: [
        { num: '۰۱', title: 'موشکافی مالی و ارزیابی ارزش ذاتی (Due Diligence)' },
        { num: '۰۲', title: 'معماری مدل کسب‌وکار و تدوین اسناد سرمایه‌گذاری' },
        { num: '۰۳', title: 'مذاکره، جذب سرمایه‌گذار و انعقاد قراردادهای سهام' },
        { num: '۰۴', title: 'استقرار هلدینگ و نظارت بر رشد بازدهی سهام' },
      ],
      sessions: 'همراهی راهبردی ۳ ماهه (۱۲ جلسه حضوری VIP + خط مستقیم)',
      duration: 'هر جلسه ۷۵ الی ۹۰ دقیقه',
      location: 'دفتر مرکزی هلدینگ / جلسات VIP اختصاصی',
      guarantee: 'عقد قرارداد رسمی عدم افشای اطلاعات تجاری (NDA)',
      actionText: 'درخواست بررسی پرونده سرمایه‌گذاری',
    },
    {
      id: 'patent-innovation',
      theme: 'sapphire',
      title: 'تجاری‌سازی نوآوری و ثبت اختراع',
      latinTitle: 'Patent Drafting & Technology Commercialization',
      tag: 'تکیه بر ۵ اختراع ثبت‌شده و دانش‌بنیان',
      icon: 'lightbulb',
      isFeatured: false,
      audience: 'مخترعان مستقل، اساتید و پژوهشگران، مراکز تحقیق و توسعه (R&D) صنایع و شرکت‌های دانش‌بنیان متقاضی صیانت و درآمدزایی از ایده.',
      description: 'مسیر مطمئن صیانت از دارایی‌های فکری و تبدیل اختراع به محصول تجاری ثروت‌آفرین. با تکیه بر تجربه عملی ثبت ۵ اختراع ملی و مدال طلای مسابقات جهانی مخترعان سوئیس.',
      features: [
        'جستجوی موشکافانه پتنت‌های بین‌المللی و ارزیابی عدم نقض حق تقدم (FTO Search)',
        'تدوین فنی و حقوقی ادعانامه‌ها (Claims) و نقشه‌های اختراع با استانداردهای رسمی',
        'پیگیری اداری و دفاع در داوری‌های علمی دانشگاهی تا اخذ گواهینامه رسمی ثبت اختراع',
        'مشاوره و هدایت برای ثبت تقاضانامه بین‌المللی تحت معاهده PCT',
        'مدل‌سازی تجاری‌سازی، ارزش‌گذاری فناوری و عقد قراردادهای واگذاری لایسنس',
        'معرفی طرح به سرمایه‌گذاران صنعتی، صندوق‌های پژوهش و فناوری و شتاب‌دهنده‌ها',
      ],
      deliverables: [
        'گزارش مکتوب تحلیل پیشینه پتنت‌های جهانی و آزادی عمل (FTO)',
        'پکیج نهایی اسناد ثبت اختراع شامل توصیف، ادعانامه و خلاصه‌نامه',
        'مدل توجیه اقتصادی و طرح تجاری‌سازی صنعتی محصول اختراعی',
      ],
      processSteps: [
        { num: '۰۱', title: 'تحلیل اختراع، جستجوی پتنت و اعتبارسنجی FTO' },
        { num: '۰۲', title: 'نگارش استاندارد ادعانامه‌ها و نقشه‌های فنی' },
        { num: '۰۳', title: 'دفاع در مراجع داوری علمی تا صدور گواهینامه رسمی' },
        { num: '۰۴', title: 'ارزش‌گذاری فناوری و تدوین مدل تجاری‌سازی صنعتی' },
      ],
      sessions: 'فرآیند مرحله‌به‌مرحله تا صدور گواهینامه و قرارداد تجاری‌سازی',
      duration: 'جلسات منظم پایش، تدوین و دفاعیات تا ثبت نهایی',
      location: 'دفتر مشاوره نوآوری و اختراعات',
      guarantee: 'حفظ ۱۰۰٪ حق تقدم و امنیت مالکیت معنوی ایده',
      actionText: 'ارزیابی اولیه طرح اختراع',
    },
    {
      id: 'management-consulting',
      theme: 'slate',
      title: 'مشاوره تخصصی مدیریت و عارضه‌یابی سازمانی',
      latinTitle: 'Specialized Management Consulting & Corporate Governance',
      tag: 'بهره‌وری، ثبات و خروج از رکود',
      icon: 'briefcase',
      isFeatured: false,
      audience: 'مدیران عامل، هیئت‌مدیره و سازمان‌هایی که با کاهش حاشیه سود، افت انگیزه سازمانی، تعارضات بین شرکا یا بن‌بست‌های اجرایی مواجهند.',
      description: 'معاینه‌ای بالینی و تخصصی بر ساختار سازمان شما. با متدولوژی عارضه‌یابی جامع، گلوگاه‌های پنهان اتلاف منابع شناسایی شده و نقشه راه ملموس برای دستیابی به بهره‌وری پایدار و سودآوری پیاده‌سازی می‌گردد.',
      features: [
        'عارضه‌یابی ۳۶۰ درجه عملکرد سازمانی، منابع انسانی و زنجیره تأمین ارزش',
        'اصلاح فرآیندهای گلوگاهی و تدوین ماتریس شاخص‌های کلیدی عملکرد (KPIs)',
        'تسهیل‌گری و حل تعارضات ساختاری میان شرکا، سهام‌داران و اعضای هیئت‌مدیره',
        'طراحی ساختار سازمانی متناسب با اهداف توسعه و تفویض اختیار شفاف',
        'تدوین آیین‌نامه‌های حاکمیت شرکتی و تدابیر حقوقی کاهش ریسک قراردادها',
        'نظارت ماهانه بر پیاده‌سازی مصوبات استراتژیک و تضمین خروجی‌های ملموس مالی',
      ],
      deliverables: [
        'گزارش تحلیلی عارضه‌یابی ۳۶۰ درجه و شناسایی گلوگاه‌های اتلاف',
        'دفترچه شاخص‌های کلیدی عملکرد (KPIs) و سیستم پاداش مبتنی بر عملکرد',
        'نقشه راه اجرایی (Action Roadmap) شش‌ماهه با تعیین مسئولیت‌ها',
      ],
      processSteps: [
        { num: '۰۱', title: 'مصاحبه‌های عمیق، تحلیل اسناد و عارضه‌یابی گلوگاه‌ها' },
        { num: '۰۲', title: 'بازطراحی فرآیندها و تدوین استراتژی اصلاحی' },
        { num: '۰۳', title: 'همراهی هیئت‌مدیره در حل تعارضات و استقرار نظام جدید' },
        { num: '۰۴', title: 'پایش مستمر شاخص‌ها و تضمین سودآوری پایدار' },
      ],
      sessions: 'دوره فشرده عارضه‌یابی و اصلاح (۶ الی ۸ جلسه اختصاصی)',
      duration: 'هر جلسه ۶۰ دقیقه + تحلیل اسناد و همراهی مدیران',
      location: 'حضوری در محل سازمان متقاضی / دفتر مشاوره دکتر طهوریان',
      guarantee: 'تضمین رازداری و ارائه راهکارهای عملیاتی و آزموده‌شده',
      actionText: 'درخواست مشاوره تخصصی مدیریت',
    },
  ],
  contacts: <Contact[]>[],
};

// Date formatting helper
function formatDate(d: Date | string): string {
  if (!d) return '';
  const date = new Date(d);
  if (isNaN(date.getTime())) return '';
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${y}/${m}/${day}`;
}

function formatDateShort(d: Date | string): string {
  return formatDate(d);
}

function enrichArticles() {
  db.articles.forEach((a) => {
    a.formattedDate = formatDate(a.created_at);
    a.formattedDateShort = formatDateShort(a.created_at);
  });
  db.categories.forEach((c) => {
    c.formattedDateShort = formatDateShort(c.created_at);
  });
}
enrichArticles();

// Sessions & CSRF
interface SessionData {
  id: string;
  authenticated: boolean;
  username: string;
  csrfToken: string;
  createdAt: number;
}
const sessions = new Map<string, SessionData>();

function toEnglishDigits(str: any): string {
  if (!str) return '';
  return String(str)
    .replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 1776))
    .replace(/[٠-٩]/g, (d) => String(d.charCodeAt(0) - 1632))
    .trim();
}

function setSessionCookie(req: Request, res: Response, sessionId: string) {
  const isHttps = req.secure || req.headers['x-forwarded-proto'] === 'https' || process.env.ENV === 'production';
  res.cookie('admin_session', sessionId, {
    httpOnly: true,
    secure: isHttps,
    sameSite: isHttps ? 'none' : 'lax',
    path: '/',
    maxAge: 30 * 24 * 60 * 60 * 1000,
  });
}

function getSessionId(req: Request): string | undefined {
  return req.cookies.admin_session || (req.query.sid as string) || (req.headers['x-admin-session'] as string);
}

function getOrCreateSession(req: Request, res: Response): SessionData {
  const sessionId = getSessionId(req);
  if (sessionId && sessions.has(sessionId)) {
    return sessions.get(sessionId)!;
  }
  const newSessionId = crypto.randomBytes(32).toString('hex');
  const session: SessionData = {
    id: newSessionId,
    authenticated: false,
    username: '',
    csrfToken: crypto.randomBytes(16).toString('hex'),
    createdAt: Date.now(),
  };
  sessions.set(newSessionId, session);
  setSessionCookie(req, res, newSessionId);
  return session;
}

// Auth middleware for admin routes
function requireAuth(req: Request, res: Response, next: NextFunction) {
  const sessionId = getSessionId(req);
  if (!sessionId || !sessions.has(sessionId)) {
    return res.redirect('/admin/login');
  }
  const session = sessions.get(sessionId)!;
  if (!session.authenticated) {
    return res.redirect('/admin/login');
  }
  // Refresh cookie on active requests
  setSessionCookie(req, res, sessionId);
  next();
}

// Contact rate limiter
const contactHits = new Map<string, number[]>();
function checkContactRateLimit(ip: string): boolean {
  const now = Date.now();
  const windowMs = 10 * 60 * 1000;
  const maxHits = 5;
  const hits = (contactHits.get(ip) || []).filter((t) => now - t < windowMs);
  if (hits.length >= maxHits) {
    contactHits.set(ip, hits);
    return false;
  }
  hits.push(now);
  contactHits.set(ip, hits);
  return true;
}

// Render helper for public pages using base layout
async function renderPage(
  res: Response,
  pageView: string,
  data: {
    title: string;
    description?: string;
    keywords?: string;
    ogImage?: string;
    ogType?: string;
    currentPath: string;
    content?: any;
  }
) {
  app.render(`pages/${pageView}`, data, (err, body) => {
    if (err) {
      console.error('Page render error:', err);
      return res.status(500).send('Render error: ' + err.message);
    }
    res.render('layout/base', {
      title: data.title,
      description: data.description,
      keywords: data.keywords,
      ogImage: data.ogImage,
      ogType: data.ogType,
      currentPath: data.currentPath,
      content: data.content,
      body,
    });
  });
}

async function renderAdmin(
  req: Request,
  res: Response,
  adminView: string,
  data: {
    title: string;
    currentPath: string;
    error?: string;
    success?: string;
    data?: any;
    csrfToken?: string;
    sid?: string;
  }
) {
  const session = getOrCreateSession(req, res);
  data.csrfToken = session.csrfToken;
  data.sid = session.id;
  data.currentPath = req.path;

  app.render(`admin/${adminView}`, { ...data, sid: session.id }, (err, body) => {
    if (err) {
      console.error('Admin page render error:', err);
      return res.status(500).send('Admin render error: ' + err.message);
    }
    res.render('admin/layout', {
      title: data.title,
      currentPath: data.currentPath,
      error: data.error,
      success: data.success,
      csrfToken: session.csrfToken,
      sid: session.id,
      body,
    });
  });
}

// -------------------------------------------------------------
// Public Routes
// -------------------------------------------------------------

app.get('/', (req, res) => {
  enrichArticles();
  const publishedArticles = db.articles.filter((a) => a.is_published);
  renderPage(res, 'index', {
    title: 'دکتر حسین طهوریان | طب سنتی و رگ‌گیری، ثبت اختراع، سرمایه‌گذاری و مشاوره تخصصی',
    description: 'وب‌سایت رسمی دکتر حسین طهوریان — خدمات تخصصی طب سنتی و رگ‌گیری دستی، ۵ اختراع ثبت‌شده رسمی، سرمایه‌گذاری هلدینگ و مشاوره تخصصی مدیریت سازمانی.',
    currentPath: '/',
    content: {
      Stats: db.stats,
      Services: db.services,
      Articles: publishedArticles,
      Gallery: db.gallery,
      GalleryCount: db.gallery.length,
    },
  });
});

app.get('/about', (req, res) => {
  renderPage(res, 'about', {
    title: 'درباره دکتر حسین طهوریان | طب سنتی، مخترع و مشاور سرمایه‌گذاری هلدینگ',
    description: 'زندگینامه و دستاوردهای دکتر حسین طهوریان؛ تخصص در طب سنتی و فنون رگ‌گیری، ۵ اختراع ثبت‌شده ملی و بین‌المللی، مدال طلای مخترعان سوئیس و راهبری هلدینگ‌ها.',
    currentPath: '/about',
    content: {
      About: db.about,
    },
  });
});

app.get('/services', (req, res) => {
  renderPage(res, 'services', {
    title: 'پکیج‌های جامع تخصصی | طب سنتی و رگ‌گیری، سرمایه‌گذاری، اختراع و مشاوره مدیریت',
    description: 'پکیج‌های جامع و اختصاصی دکتر حسین طهوریان: طب سنتی و رگ‌گیری تخصصی دستی، توسعه سرمایه‌گذاری و هلدینگ، تجاری‌سازی نوآوری و ثبت اختراع، و مشاوره تخصصی مدیریت با تضمین محرمانگی NDA.',
    currentPath: '/services',
    content: {
      Services: db.services,
      Packages: db.packages,
    },
  });
});

app.get('/projects', (req, res) => {
  renderPage(res, 'projects', {
    title: 'پروژه‌ها و هلدینگ‌های همکار | دکتر حسین طهوریان',
    description: 'هلدینگ سرآمد سرمایه ایلیا، گروه پتنت و پروژه‌های موفق سرمایه‌گذاری و دانش‌بنیان دکتر طهوریان.',
    currentPath: '/projects',
    content: {
      Projects: db.gallery,
    },
  });
});

app.get('/articles', (req, res) => {
  enrichArticles();
  const publishedArticles = db.articles.filter((a) => a.is_published);
  const categoryCounts: Record<string, number> = {};
  publishedArticles.forEach((a) => {
    const cat = (a.category || '').trim();
    if (cat) categoryCounts[cat] = (categoryCounts[cat] || 0) + 1;
  });

  const categories = db.categories.map((c) => ({
    slug: c.slug,
    title: c.title,
    count: categoryCounts[c.slug] || 0,
  }));

  const categoryTitles: Record<string, string> = {};
  db.categories.forEach((c) => {
    categoryTitles[c.slug] = c.title;
  });

  renderPage(res, 'articles', {
    title: 'مقالات تخصصی و یادداشت‌های علمی | دکتر حسین طهوریان',
    description: 'مجموعه مقالات تخصصی دکتر حسین طهوریان پیرامون آموزش ثبت اختراع، تسکین دردهای عضلانی با رگ‌گیری و مبانی اصیل طب سنتی ایرانی و مزاج‌شناسی.',
    keywords: 'مقالات دکتر طهوریان, آموزش ثبت اختراع, تسکین دردهای عضلانی, رگ گیری, طب سنتی ایرانی, مزاج شناسی, پتنت, درمان دستی',
    ogImage: '/static/images/download%20(1)-1.webp',
    ogType: 'website',
    currentPath: '/articles',
    content: {
      Articles: publishedArticles,
      Categories: categories,
      CategoryTitles: categoryTitles,
    },
  });
});

app.get('/articles/:slug', (req, res) => {
  enrichArticles();
  const slug = req.params.slug;
  const article = db.articles.find((a) => a.slug === slug);
  if (!article) {
    return res.redirect('/articles');
  }
  const publishedArticles = db.articles.filter((a) => a.is_published && a.slug !== slug);
  const recentArticles = publishedArticles.slice(0, 4);

  renderPage(res, 'article_detail', {
    title: `${article.title} | دکتر حسین طهوریان`,
    description: article.summary,
    keywords: article.keywords,
    ogImage: article.image_url,
    ogType: 'article',
    currentPath: `/articles/${slug}`,
    content: {
      Article: article,
      RecentArticles: recentArticles,
    },
  });
});

app.get('/team', (req, res) => {
  renderPage(res, 'team', {
    title: 'تیم ما | طهوریان',
    description: 'تیم متخصص هلدینگ طهوریان',
    currentPath: '/team',
  });
});

app.get('/contact', (req, res) => {
  renderPage(res, 'contact', {
    title: 'تماس با ما | دکتر حسین طهوریان',
    description: 'راه‌های ارتباطی با هلدینگ طهوریان',
    currentPath: '/contact',
    content: db.contact_info,
  });
});

app.post('/api/contact', (req, res) => {
  const ip = req.ip || req.socket.remoteAddress || '127.0.0.1';
  const isHtmx = req.headers['hx-request'] === 'true';

  if (!checkContactRateLimit(ip)) {
    if (isHtmx) {
      res.setHeader('HX-Trigger', JSON.stringify({
        contactToast: {
          type: 'error',
          title: 'تعداد درخواست‌ها زیاد شد',
          message: 'برای جلوگیری از سوءاستفاده، ارسال محدود شده است. لطفاً چند دقیقه بعد دوباره تلاش کنید.',
        },
      }));
      return res.status(429).send(
        '<div class="contact-response error"><i data-lucide="alert-circle"></i><div><strong>تعداد درخواست‌ها زیاد شد</strong><p>برای جلوگیری از سوءاستفاده، ارسال محدود شده است. لطفاً چند دقیقه بعد دوباره تلاش کنید.</p></div></div>'
      );
    }
    return res.status(429).json({ status: 'error', message: 'تعداد درخواست‌ها زیاد شد؛ لطفاً کمی بعد تلاش کنید' });
  }

  const { name, email, phone, subject, message } = req.body;
  if (!name || !message) {
    if (isHtmx) {
      res.setHeader('HX-Trigger', JSON.stringify({
        contactToast: {
          type: 'error',
          title: 'خطا در دریافت اطلاعات',
          message: 'لطفاً فیلدهای ضروری را به درستی تکمیل فرمایید.',
        },
      }));
      return res.status(400).send(
        '<div class="contact-response error"><i data-lucide="alert-circle"></i><div><strong>خطا در دریافت اطلاعات</strong><p>لطفاً فیلدهای ضروری را به درستی تکمیل فرمایید.</p></div></div>'
      );
    }
    return res.status(400).json({ status: 'error', message: 'نام و متن پیام الزامی است' });
  }

  const contact: Contact = {
    id: nextContactId++,
    name: String(name).trim(),
    email: String(email || '').trim(),
    phone: String(phone || '').trim(),
    subject: String(subject || '').trim(),
    message: String(message).trim(),
    created_at: new Date(),
  };
  db.contacts.push(contact);

  if (isHtmx) {
    res.setHeader('HX-Trigger', JSON.stringify({
      contactToast: {
        type: 'success',
        title: 'پیام شما با موفقیت ثبت شد',
        message: 'درخواست شما دریافت شد و در کمتر از ۲۴ ساعت کاری با شما تماس خواهیم گرفت.',
      },
    }));
    return res.send(
      '<div class="contact-response success"><i data-lucide="check-circle-2"></i><div><strong>پیام شما با موفقیت ثبت شد</strong><p>درخواست شما دریافت شد و در کمتر از ۲۴ ساعت کاری با شما تماس خواهیم گرفت.</p></div></div>'
    );
  }

  return res.json({ status: 'success', message: 'پیام شما با موفقیت ارسال شد' });
});

// -------------------------------------------------------------
// Admin Authentication Routes (Hardened Security)
// -------------------------------------------------------------

interface LoginAttempt {
  count: number;
  lockedUntil: number;
}
const loginAttempts = new Map<string, LoginAttempt>();

function isLoginRateLimited(ip: string): boolean {
  const now = Date.now();
  const attempt = loginAttempts.get(ip);
  if (!attempt) return false;
  if (attempt.lockedUntil > now) return true;
  if (now - attempt.lockedUntil > 15 * 60 * 1000) {
    loginAttempts.delete(ip);
    return false;
  }
  return false;
}

function recordFailedLogin(ip: string): number {
  const now = Date.now();
  const attempt = loginAttempts.get(ip) || { count: 0, lockedUntil: 0 };
  attempt.count++;
  if (attempt.count >= 5) {
    attempt.lockedUntil = now + 15 * 60 * 1000; // Lock for 15 minutes after 5 failed attempts
  }
  loginAttempts.set(ip, attempt);
  return attempt.count;
}

function clearLoginAttempts(ip: string) {
  loginAttempts.delete(ip);
}

function timingSafeEqualStr(a: string, b: string): boolean {
  const hashA = crypto.createHash('sha256').update(String(a || '')).digest();
  const hashB = crypto.createHash('sha256').update(String(b || '')).digest();
  return crypto.timingSafeEqual(hashA, hashB);
}

app.get('/admin/login', (req, res) => {
  const sessionId = getSessionId(req);
  if (sessionId && sessions.has(sessionId) && sessions.get(sessionId)!.authenticated) {
    return res.redirect('/admin/dashboard?sid=' + sessionId);
  }
  res.render('admin/login', {
    title: 'ورود به پنل مدیریت | دکتر حسین طهوریان',
    error: '',
  });
});

app.post('/admin/login', async (req, res) => {
  const ip = req.ip || req.socket.remoteAddress || '127.0.0.1';

  const rawUser = req.body.username;
  const rawPass = req.body.password;

  const normalizedUser = toEnglishDigits(rawUser).toLowerCase();
  const normalizedPass = toEnglishDigits(rawPass);

  // Check valid passwords (support 09152491524, ADMIN_PASS, etc.)
  const isPassValid =
    normalizedPass === '09152491524' ||
    normalizedPass === toEnglishDigits(ADMIN_PASS) ||
    timingSafeEqualStr(normalizedPass, '09152491524') ||
    timingSafeEqualStr(normalizedPass, ADMIN_PASS);

  if (isPassValid) {
    clearLoginAttempts(ip);

    // Invalidate old session ID
    const oldSessionId = getSessionId(req);
    if (oldSessionId) {
      sessions.delete(oldSessionId);
    }

    // Create fresh authenticated session
    const newSessionId = crypto.randomBytes(32).toString('hex');
    const session: SessionData = {
      id: newSessionId,
      authenticated: true,
      username: normalizedUser || 'admin',
      csrfToken: crypto.randomBytes(16).toString('hex'),
      createdAt: Date.now(),
    };
    sessions.set(newSessionId, session);
    setSessionCookie(req, res, newSessionId);

    return res.redirect('/admin/dashboard?sid=' + newSessionId);
  }

  // If password failed, check rate limiting
  if (isLoginRateLimited(ip)) {
    const attempt = loginAttempts.get(ip);
    const remainingMinutes = attempt ? Math.max(1, Math.ceil((attempt.lockedUntil - Date.now()) / 60000)) : 15;
    return res.status(429).render('admin/login', {
      title: 'ورود به پنل مدیریت',
      error: `به دلیل تلاش‌های ناموفق مکرر، ورود به پنل برای ${remainingMinutes} دقیقه مسدود شده است.`,
    });
  }

  const failedCount = recordFailedLogin(ip);
  const remaining = Math.max(0, 5 - failedCount);
  const errorMsg = remaining > 0
    ? `رمز عبور وارد شده نادرست است. (${remaining} بار تلاش باقیمانده)`
    : 'به دلیل ۵ بار تلاش ناموفق، حساب برای ۱۵ دقیقه مسدود شد.';

  res.render('admin/login', {
    title: 'ورود به پنل مدیریت',
    error: errorMsg,
  });
});

app.get('/admin/logout', (req, res) => {
  const sessionId = getSessionId(req);
  if (sessionId) {
    sessions.delete(sessionId);
  }
  res.clearCookie('admin_session');
  res.redirect('/admin/login');
});

// -------------------------------------------------------------
// Admin Protected Routes
// -------------------------------------------------------------

app.get('/admin', requireAuth, (req, res) => {
  res.redirect('/admin/dashboard');
});

app.get('/admin/dashboard', requireAuth, (req, res) => {
  enrichArticles();
  const publishedCount = db.articles.filter((a) => a.is_published).length;
  renderAdmin(req, res, 'dashboard', {
    title: 'داشبورد مدیریت | طهوریان',
    currentPath: '/admin/dashboard',
    data: {
      TotalArticles: db.articles.length,
      PublishedArticles: publishedCount,
      DraftArticles: db.articles.length - publishedCount,
      RecentArticles: db.articles.slice(0, 5),
      ActiveSections: 1,
      TotalSections: 5,
    },
  });
});

app.get('/admin/sections', requireAuth, (req, res) => {
  renderAdmin(req, res, 'sections', {
    title: 'مدیریت بخش‌های سایت | طهوریان',
    currentPath: '/admin/sections',
    data: {
      ArticleCount: db.articles.length,
    },
  });
});

app.get('/admin/articles', requireAuth, (req, res) => {
  enrichArticles();
  const msg = req.query.msg;
  let successMsg = '';
  if (msg === 'created') successMsg = 'مقاله جدید با موفقیت ایجاد شد.';
  if (msg === 'updated') successMsg = 'مقاله با موفقیت ویرایش شد.';
  if (msg === 'deleted') successMsg = 'مقاله با موفقیت حذف شد.';

  renderAdmin(req, res, 'articles_list', {
    title: 'مدیریت مقالات | طهوریان',
    currentPath: '/admin/articles',
    success: successMsg,
    data: {
      Articles: db.articles,
    },
  });
});

app.get('/admin/articles/new', requireAuth, (req, res) => {
  renderAdmin(req, res, 'article_form', {
    title: 'افزودن مقاله جدید | طهوریان',
    currentPath: '/admin/articles/new',
    data: {
      IsNew: true,
      Article: {
        author: 'دکتر حسین طهوریان',
        is_published: true,
      },
      Categories: db.categories,
    },
  });
});

app.post('/admin/articles/new', requireAuth, (req, res) => {
  const { title, slug, category, summary, content, image_url, author, is_published } = req.body;
  if (!title || !title.trim()) {
    return renderAdmin(req, res, 'article_form', {
      title: 'افزودن مقاله جدید | طهوریان',
      currentPath: '/admin/articles/new',
      error: 'عنوان مقاله نمی‌تواند خالی باشد.',
      data: {
        IsNew: true,
        Article: req.body,
        Categories: db.categories,
      },
    });
  }

  const generatedSlug = (slug && slug.trim())
    ? slug.trim().toLowerCase().replace(/\s+/g, '-')
    : title.trim().toLowerCase().replace(/\s+/g, '-').replace(/[^\w\u0600-\u06FF-]+/g, '') || `article-${Date.now()}`;

  const newArticle: Article = {
    id: nextArticleId++,
    title: title.trim(),
    slug: generatedSlug,
    category: (category || '').trim(),
    summary: (summary || '').trim(),
    content: (content || '').trim(),
    image_url: (image_url || '').trim(),
    author: (author || 'دکتر حسین طهوریان').trim(),
    is_published: is_published === 'true' || is_published === 'on' || is_published === '1',
    created_at: new Date(),
    updated_at: new Date(),
  };

  db.articles.unshift(newArticle);
  res.redirect('/admin/articles?msg=created');
});

app.get('/admin/articles/edit/:id', requireAuth, (req, res) => {
  const id = parseInt(req.params.id, 10);
  const article = db.articles.find((a) => a.id === id);
  if (!article) {
    return res.redirect('/admin/articles');
  }

  renderAdmin(req, res, 'article_form', {
    title: `ویرایش مقاله: ${article.title}`,
    currentPath: '/admin/articles',
    data: {
      IsNew: false,
      Article: article,
      Categories: db.categories,
    },
  });
});

app.post('/admin/articles/edit/:id', requireAuth, (req, res) => {
  const id = parseInt(req.params.id, 10);
  const article = db.articles.find((a) => a.id === id);
  if (!article) {
    return res.redirect('/admin/articles');
  }

  const { title, slug, category, summary, content, image_url, author, is_published } = req.body;
  if (!title || !title.trim()) {
    return renderAdmin(req, res, 'article_form', {
      title: 'ویرایش مقاله',
      currentPath: '/admin/articles',
      error: 'عنوان مقاله نمی‌تواند خالی باشد.',
      data: {
        IsNew: false,
        Article: { ...article, ...req.body },
        Categories: db.categories,
      },
    });
  }

  article.title = title.trim();
  if (slug && slug.trim()) article.slug = slug.trim();
  article.category = (category || '').trim();
  article.summary = (summary || '').trim();
  article.content = (content || '').trim();
  article.image_url = (image_url || '').trim();
  article.author = (author || 'دکتر حسین طهوریان').trim();
  article.is_published = is_published === 'true' || is_published === 'on' || is_published === '1';
  article.updated_at = new Date();

  res.redirect('/admin/articles?msg=updated');
});

app.post('/admin/articles/delete/:id', requireAuth, (req, res) => {
  const id = parseInt(req.params.id, 10);
  const idx = db.articles.findIndex((a) => a.id === id);
  if (idx !== -1) {
    db.articles.splice(idx, 1);
  }
  res.redirect('/admin/articles?msg=deleted');
});

app.get('/admin/categories', requireAuth, (req, res) => {
  enrichArticles();
  const msg = req.query.msg;
  let successMsg = '';
  if (msg === 'created') successMsg = 'دسته‌بندی جدید با موفقیت ایجاد شد.';
  if (msg === 'deleted') successMsg = 'دسته‌بندی حذف شد.';

  renderAdmin(req, res, 'categories', {
    title: 'دسته‌بندی مقالات | طهوریان',
    currentPath: '/admin/categories',
    success: successMsg,
    data: {
      Categories: db.categories,
    },
  });
});

app.post('/admin/categories/new', requireAuth, (req, res) => {
  const { title, slug } = req.body;
  if (!title || !title.trim()) {
    return renderAdmin(req, res, 'categories', {
      title: 'دسته‌بندی مقالات | طهوریان',
      currentPath: '/admin/categories',
      error: 'عنوان دسته‌بندی الزامی است.',
      data: {
        Categories: db.categories,
        FormTitle: title,
        FormSlug: slug,
      },
    });
  }

  const generatedSlug = (slug && slug.trim())
    ? slug.trim().toLowerCase().replace(/\s+/g, '-')
    : title.trim().toLowerCase().replace(/\s+/g, '-').replace(/[^\w\u0600-\u06FF-]+/g, '') || `cat-${Date.now()}`;

  const newCat: ArticleCategory = {
    id: nextCategoryId++,
    title: title.trim(),
    slug: generatedSlug,
    created_at: new Date(),
  };

  db.categories.push(newCat);
  res.redirect('/admin/categories?msg=created');
});

app.post('/admin/categories/delete/:id', requireAuth, (req, res) => {
  const id = parseInt(req.params.id, 10);
  const idx = db.categories.findIndex((c) => c.id === id);
  if (idx !== -1) {
    db.categories.splice(idx, 1);
  }
  res.redirect('/admin/categories?msg=deleted');
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).send('ok');
});

// Start listening
app.listen(PORT, HOST, () => {
  console.log(`Server listening on http://${HOST}:${PORT} in ${process.env.ENV || 'development'} mode`);
});
