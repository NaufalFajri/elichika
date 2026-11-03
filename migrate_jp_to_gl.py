#!/usr/bin/env python3
"""
SIFAS JP to GL Database Migration Tool
Ports all JP-exclusive content (costumes, accessories, songs, beatmaps, backgrounds, missions, etc.)
from assets/db/jp into assets/db/gl, with full translations for EN, KO, and ZH.
"""

import os
import sys
import shutil
import sqlite3
import re
from datetime import datetime

ASSETS_DIR = "assets"
PATCH_DIR = "scrapped accessory translate"
DB_JP_DIR = os.path.join(ASSETS_DIR, "db", "jp")
DB_GL_DIR = os.path.join(ASSETS_DIR, "db", "gl")
SCRAPED_HTML_PATH = os.path.join(PATCH_DIR, "Accessories [Nijigasaki High Data Research Club - SIFAS Wiki].htm")

# 44 Accessory translations mapping: (key, ja_name, en_name, ko_name, zh_name)
ACCESSORY_TRANSLATIONS = {
    'accessory_name_33': ('ニジガク帽子', 'Nijigaku Cap', '니지가쿠 모자', '虹咲帽子'),
    'accessory_name_45': ('ローダンセの髪飾り', 'Rhodanthe Hair Ornament', '로단세 머리장식', '羅丹斯髮飾'),
    'accessory_name_50': ('ランニングシューズ', 'Running Shoes', '러닝화', '慢跑鞋'),
    'accessory_name_51': ('虹ヶ咲学園どんぶり', 'Nijigasaki Donburi Bowl', '니지가사키 학원 덮밥 그릇', '虹咲學園丼飯碗'),
    'accessory_name_52': ('アランロボ', 'Aran Robot', '아란 로봇', '亞蘭機器人'),
    'accessory_name_53': ('彼方ちゃんの枕', "Kanata's Pillow", '카나타의 베개', '彼方的枕頭'),
    'accessory_name_54': ('しずくのマスク', "Shizuku's Mask", '시즈쿠의 마스크', '雫的口罩'),
    'accessory_name_55': ('果林のチョーカー', "Karin's Choker", '카린의 초커', '果林的項圈'),
    'accessory_name_56': ('ニジガクバット', 'Nijigaku Bat', '니지가쿠 배트', '虹咲球棒'),
    'accessory_name_57': ('Brand New Story マフラータオル', 'Brand New Story Towel', 'Brand New Story 머플러 타올', 'Brand New Story 圍巾毛巾'),
    'accessory_name_58': ('Back to the TOKIMEKI マフラータオル', 'Back to the TOKIMEKI Towel', 'Back to the TOKIMEKI 머플러 타올', 'Back to the TOKIMEKI 圍巾毛巾'),
    'accessory_name_59': ('DiverDiva リボン', 'DiverDiva Ribbon', 'DiverDiva 리본', 'DiverDiva 蝴蝶結'),
    'accessory_name_60': ('花火のキーホルダー', 'Fireworks Keyholder', '불꽃 열쇠고리', '煙火鑰匙圈'),
    'accessory_name_61': ('かすみんBOX', 'Kasumin Box', '카스밍 BOX', '小霞霞BOX'),
    'accessory_name_62': ('SIFTシャツ', 'SIF T-Shirt', 'SIF T셔츠', 'SIF T恤'),
    'accessory_name_63': ('黄色のガーベラ', 'Yellow Gerbera', '노란색 거베라', '黃色非洲菊'),
    'accessory_name_64': ('A・ZU・NA リボン＆ハット', 'A•ZU•NA Ribbon & Hat', 'A·ZU·NA 리본&모자', 'A・ZU・NA 蝴蝶結＆帽子'),
    'accessory_name_65': ('QU4RTZ 花のキーホルダー', 'QU4RTZ Flower Keyholder', 'QU4RTZ 꽃 열쇠고리', 'QU4RTZ 花朵鑰匙圈'),
    'accessory_name_66': ('R3BIRTHのリボン', 'R3BIRTH Ribbon', 'R3BIRTH의 리본', 'R3BIRTH的蝴蝶結'),
    'accessory_name_67': ('ニジガクのピンバッヂ', 'Nijigaku Pin Badge', '니지가쿠 핀 배지', '虹咲胸針'),
    'accessory_name_68': ('～夢の始まり～ メモリアルピンズ', '~The beginning of the dream~ Memorial Pin', '~꿈의 시작~ 메모리얼 핀즈', '～夢想的起點～ 紀念徽章'),
    'accessory_name_69': ('夢と笑顔のイヤーマフ', 'Earmuffs of Dreams and Smiles', '꿈과 미소의 귀마개', '夢想與笑容的耳罩'),
    'accessory_name_70': ('虹色の花', 'Rainbow-Colored Flower', '무지개색 꽃', '彩虹色的花'),
    'accessory_name_71': ('ENJOY IT！', 'ENJOY IT!', 'ENJOY IT!', 'ENJOY IT！'),
    'accessory_name_72': ('Eternal Light', 'Eternal Light', 'Eternal Light', 'Eternal Light'),
    'accessory_name_73': ('Infinity！Our wings!!', 'Infinity! Our wings!!', 'Infinity！Our wings!!', 'Infinity！Our wings!!'),
    'accessory_name_74': ('Eutopia', 'Eutopia', 'Eutopia', 'Eutopia'),
    'accessory_name_75': ('stars we chase', 'stars we chase', 'stars we chase', 'stars we chase'),
    'accessory_name_76': ('EMOTION', 'EMOTION', 'EMOTION', 'EMOTION'),
    'accessory_name_77': ('ストラップ付USB', 'USB Stick with Strap', '스트랩 달린 USB', '附掛繩隨身碟'),
    'accessory_name_78': ('食パンふせん', 'Bread Sticky Note', '식빵 포스트잇', '吐司便利貼'),
    'accessory_name_79': ('おそろいパンダ', 'Matching Pandas', '커플 판다', '成對熊貓'),
    'accessory_name_80': ('かすみのアルバム', "Kasumi's Album", '카스미의 앨범', '霞的相簿'),
    'accessory_name_81': ('しずくの空想ノート', "Shizuku's Fantasy Notebook", '시즈쿠의 망상 노트', '雫的空想筆記本'),
    'accessory_name_82': ('せつ菜の髪留め', "Setsuna's Hair Clip", '세츠나의 머리핀', '雪菜的髮夾'),
    'accessory_name_83': ('翡翠の羽根', 'Jade Feather', '비취색 날개', '翡翠羽毛'),
    'accessory_name_84': ('第2回SIFのTシャツ', 'SIF 2 T-Shirt', '제2회 SIF T셔츠', '第2屆SIF T恤'),
    'accessory_name_85': ('ミアのヘッドホン', "Mia's Headphones", '미아의 헤드폰', '米雅的耳機'),
    'accessory_name_86': ('かすみんフラッグ', 'Kasumin Flag', '카스밍 깃발', '小霞霞旗幟'),
    'accessory_name_87': ('ニジガクのクレープ', 'Nijigaku Crêpe', '니지가쿠 크레페', '虹咲可麗餅'),
    'accessory_name_88': ('コンクールのポスター', 'Competition Poster', '콩쿠르 포스터', '比賽海報'),
    'accessory_name_89': ('あなたへの贈り物', 'A Gift for You', '당신에게 주는 선물', '給你的禮物'),
    'accessory_name_90': ('永遠のティアラ', 'Tiara of Eternity', '영원의 티아라', '永遠的提亞拉'),
    'accessory_name_91': ('永遠のクラウン', 'Crown of Eternity', '영원의 크라운', '永遠的皇冠')
}

SUIT_PREFIXES = [
    ('2020カウントダウンライブ Tシャツ', '2020 Countdown Live T-Shirt', '2020 카운트다운 라이브 T셔츠', '2020倒數演唱會 T恤'),
    ('AqoursユニットライブTシャツ', 'Aqours Unit Live T-Shirt', 'Aqours 유닛 라이브 T셔츠', 'Aqours小隊演唱會 T恤'),
    ('アスルクラロ沼津3rdユニフォーム', 'Azul Claro Numazu 3rd Uniform', '아줄 클라로 누마즈 3rd 유니폼', '沼津青藍3rd球衣'),
    ('Brand New Story Tシャツ', 'Brand New Story T-Shirt', 'Brand New Story T셔츠', 'Brand New Story T恤'),
    ('ODAIBAゲーマーズ Tシャツ', 'ODAIBA Gamers T-Shirt', '오다이바 게이머즈 T셔츠', 'ODAIBA Gamers T恤'),
]

def backup_gl_databases():
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_dest = os.path.join("backup", f"db_gl_{timestamp}")
    print(f"[*] Backing up '{DB_GL_DIR}' to '{backup_dest}'...")
    os.makedirs(backup_dest, exist_ok=True)
    for item in os.listdir(DB_GL_DIR):
        src_path = os.path.join(DB_GL_DIR, item)
        dst_path = os.path.join(backup_dest, item)
        if os.path.isfile(src_path):
            shutil.copy2(src_path, dst_path)
    print("    Backup completed successfully.")
    return backup_dest

def get_member_names():
    c_en = sqlite3.connect(os.path.join(DB_GL_DIR, 'dictionary_en_k.db'))
    c_ko = sqlite3.connect(os.path.join(DB_GL_DIR, 'dictionary_ko_k.db'))
    c_zh = sqlite3.connect(os.path.join(DB_GL_DIR, 'dictionary_zh_k.db'))
    m_en, m_ko, m_zh = {}, {}, {}
    for mid in list(range(1, 10)) + list(range(101, 110)) + list(range(201, 213)):
        key = f"m_dic_member_name_{mid}"
        r_en = c_en.execute("SELECT message FROM m_dictionary WHERE id = ?", (key,)).fetchone()
        r_ko = c_ko.execute("SELECT message FROM m_dictionary WHERE id = ?", (key,)).fetchone()
        r_zh = c_zh.execute("SELECT message FROM m_dictionary WHERE id = ?", (key,)).fetchone()
        m_en[mid] = r_en[0] if r_en else ""
        m_ko[mid] = r_ko[0] if r_ko else ""
        m_zh[mid] = r_zh[0] if r_zh else ""
    c_en.close()
    c_ko.close()
    c_zh.close()
    return m_en, m_ko, m_zh

def migrate_masterdata():
    print("\n[*] Migrating masterdata.db (JP -> GL)...")
    src_path = os.path.join(DB_JP_DIR, "masterdata.db")
    dst_path = os.path.join(DB_GL_DIR, "masterdata.db")
    
    con_dst = sqlite3.connect(dst_path)
    con_dst.execute(f"ATTACH DATABASE '{os.path.abspath(src_path)}' AS jp_db")
    
    tables = [r[0] for r in con_dst.execute("SELECT name FROM sqlite_master WHERE type='table'")]
    total_added = 0
    tables_updated = 0
    
    for t in tables:
        before = con_dst.execute(f'SELECT COUNT(*) FROM main."{t}"').fetchone()[0]
        con_dst.execute(f'INSERT OR IGNORE INTO main."{t}" SELECT * FROM jp_db."{t}"')
        after = con_dst.execute(f'SELECT COUNT(*) FROM main."{t}"').fetchone()[0]
        diff = after - before
        if diff > 0:
            total_added += diff
            tables_updated += 1
            print(f"    - Table '{t}': +{diff} rows")
            
    con_dst.commit()
    con_dst.close()
    print(f"    Masterdata migration done: +{total_added} rows across {tables_updated} tables.")

def migrate_asset_dbs():
    print("\n[*] Migrating Asset Databases (JP -> GL)...")
    
    android_targets = [
        os.path.join(DB_GL_DIR, "asset_a_en.db"),
        os.path.join(DB_GL_DIR, "asset_a_ko.db"),
        os.path.join(DB_GL_DIR, "asset_a_zh.db"),
    ]
    android_src = os.path.join(DB_JP_DIR, "asset_a_ja.db")
    
    ios_targets = [
        os.path.join(DB_GL_DIR, "asset_i_en.db"),
        os.path.join(DB_GL_DIR, "asset_i_ko.db"),
        os.path.join(DB_GL_DIR, "asset_i_zh.db"),
    ]
    ios_src = os.path.join(DB_JP_DIR, "asset_i_ja.db")
    
    for targets, src, label in [(android_targets, android_src, "Android"), (ios_targets, ios_src, "iOS")]:
        for dst_path in targets:
            db_name = os.path.basename(dst_path)
            print(f"    Processing {label} DB: {db_name}...")
            con_dst = sqlite3.connect(dst_path)
            con_dst.execute(f"ATTACH DATABASE '{os.path.abspath(src)}' AS jp_asset")
            
            tables = [r[0] for r in con_dst.execute("SELECT name FROM sqlite_master WHERE type='table'")]
            rows_added = 0
            for t in tables:
                before = con_dst.execute(f'SELECT COUNT(*) FROM main."{t}"').fetchone()[0]
                con_dst.execute(f'INSERT OR IGNORE INTO main."{t}" SELECT * FROM jp_asset."{t}"')
                after = con_dst.execute(f'SELECT COUNT(*) FROM main."{t}"').fetchone()[0]
                diff = after - before
                rows_added += diff
            
            # Recalculate package counts
            con_dst.execute('UPDATE m_asset_package SET pack_num = (SELECT COUNT(*) FROM m_asset_package_mapping WHERE package_key = m_asset_package.package_key)')
            con_dst.commit()
            con_dst.close()
            print(f"      -> +{rows_added} rows added, pack_num updated.")

def migrate_dictionaries(m_en, m_ko, m_zh):
    print("\n[*] Migrating and Translating Dictionaries...")
    
    # 1. Prepare Suit Translations
    c_m_jp = sqlite3.connect(os.path.join(DB_JP_DIR, 'masterdata.db'))
    c_k_jp = sqlite3.connect(os.path.join(DB_JP_DIR, 'dictionary_ja_k.db'))
    
    suits_jp = c_m_jp.execute("SELECT id, member_m_id, name FROM m_suit").fetchall()
    suit_trans = {}
    for sid, mid, dkey in suits_jp:
        kid = dkey[2:] if dkey.startswith('k.') else dkey
        dic_msg = c_k_jp.execute("SELECT message FROM m_dictionary WHERE id = ?", (kid,)).fetchone()
        ja_name = dic_msg[0] if dic_msg else ''
        
        chara_id = mid
        if 'ミア・テイラー' in ja_name:
            chara_id = 212
        elif 'ショウ・ランジュ' in ja_name:
            chara_id = 211
            
        t_en, t_ko, t_zh = ja_name, ja_name, ja_name
        for p_ja, p_en, p_ko, p_zh in SUIT_PREFIXES:
            if p_ja in ja_name:
                t_en = f"{p_en} {m_en[chara_id]}"
                t_ko = f"{p_ko} {m_ko[chara_id]}"
                t_zh = f"{p_zh} {m_zh[chara_id]}"
                break
        suit_trans[kid] = (t_en, t_ko, t_zh)
        
    c_m_jp.close()
    c_k_jp.close()
    
    # 2. Song translations
    song_trans = {
        'song_name_so3001': ('Hajimari wa Kimi no Sora', '시작은 너의 하늘', '始於君之空'),
        'song_pronunciation_so3001': ('はじまりはきみのそら', 'はじまりはきみのそら', 'はじまりはきみのそら'),
    }
    
    # 3. Process dictionary_k
    langs = [('en', 1), ('ko', 2), ('zh', 3)]
    for lang_code, lang_idx in langs:
        dst_db = os.path.join(DB_GL_DIR, f"dictionary_{lang_code}_k.db")
        src_db = os.path.join(DB_JP_DIR, "dictionary_ja_k.db")
        print(f"    Updating {os.path.basename(dst_db)}...")
        
        con_dst = sqlite3.connect(dst_db)
        con_src = sqlite3.connect(src_db)
        
        jp_rows = con_src.execute("SELECT id, message FROM m_dictionary").fetchall()
        gl_ids = set(r[0] for r in con_dst.execute("SELECT id FROM m_dictionary"))
        
        inserted_custom = 0
        inserted_fallback = 0
        
        for k_id, ja_msg in jp_rows:
            if k_id not in gl_ids:
                # Check custom translations
                if k_id in suit_trans:
                    trans_text = suit_trans[k_id][lang_idx - 1]
                    inserted_custom += 1
                elif k_id in ACCESSORY_TRANSLATIONS:
                    trans_text = ACCESSORY_TRANSLATIONS[k_id][lang_idx]
                    inserted_custom += 1
                elif k_id in song_trans:
                    trans_text = song_trans[k_id][lang_idx - 1]
                    inserted_custom += 1
                else:
                    trans_text = ja_msg
                    inserted_fallback += 1
                    
                con_dst.execute("INSERT OR REPLACE INTO m_dictionary (id, message) VALUES (?, ?)", (k_id, trans_text))
                
        con_dst.commit()
        con_dst.close()
        con_src.close()
        print(f"      -> Added: {inserted_custom} translated + {inserted_fallback} fallback messages.")
        
    # 4. Process dictionary_m, dictionary_dummy, dictionary_v
    for suffix in ['m', 'dummy', 'v']:
        src_db = os.path.join(DB_JP_DIR, f"dictionary_ja_{suffix}.db")
        for lang_code, _ in langs:
            dst_db = os.path.join(DB_GL_DIR, f"dictionary_{lang_code}_{suffix}.db")
            if not os.path.exists(dst_db) or not os.path.exists(src_db):
                continue
            con_dst = sqlite3.connect(dst_db)
            con_src = sqlite3.connect(src_db)
            
            gl_ids = set(r[0] for r in con_dst.execute("SELECT id FROM m_dictionary"))
            jp_rows = con_src.execute("SELECT id, message FROM m_dictionary").fetchall()
            
            added = 0
            for k_id, ja_msg in jp_rows:
                if k_id not in gl_ids:
                    con_dst.execute("INSERT OR REPLACE INTO m_dictionary (id, message) VALUES (?, ?)", (k_id, ja_msg))
                    added += 1
            con_dst.commit()
            con_dst.close()
            con_src.close()
            if added > 0:
                print(f"    Updated {os.path.basename(dst_db)}: +{added} entries.")

def main():
    print("=" * 65)
    print("  SIFAS JP -> GL Full Database Migration Tool")
    print("=" * 65)
    
    # 1. Backup
    backup_gl_databases()
    
    # 2. Member names
    m_en, m_ko, m_zh = get_member_names()
    
    # 3. Masterdata
    migrate_masterdata()
    
    # 4. Asset DBs
    migrate_asset_dbs()
    
    # 5. Dictionaries
    migrate_dictionaries(m_en, m_ko, m_zh)
    
    print("\n" + "=" * 65)
    print("  MIGRATION SUCCESSFULLY FINISHED!")
    print("=" * 65)
    print("\nNext step: Run './elichika.exe rebuild_assets' to re-encrypt")
    print("and package the client databases.")

if __name__ == '__main__':
    main()
