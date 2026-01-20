import { useEffect, useState, useCallback } from "react";
import {
  View,
  Text,
  TextInput,
  ScrollView,
  StyleSheet,
  Image,
  TouchableOpacity,
  ActivityIndicator,
  Platform,
} from "react-native";

const BACKEND_URL = "http://localhost:8080";

// Theme Colors
const COLORS = {
  background: "#000000",
  primary: "#39FF14", // Neon Green
  text: "#FFFFFF",
  glass: "rgba(255, 255, 255, 0.05)",
  glassBorder: "rgba(255, 255, 255, 0.5)", // Increased visibility
  gold: "#FFD700",
  silver: "#C0C0C0",
  bronze: "#CD7F32",
};

const PODIUM_IMAGES: Record<number, string> = {
  1: "https://static.vecteezy.com/system/resources/thumbnails/054/555/561/small/a-man-wearing-headphones-and-sunglasses-is-wearing-a-hoodie-free-vector.jpg",
  2: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRYTAxdI7vxMZivc6xKNu8lmtIG6-9tAWUSbqX90RIneA&s",
  3: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQKh_EdrcFNluYTeHnoB3f-juV7lTs82h5OBLdVrnVplw&s"
};

const LOGO_URI = "data:image/png;base64,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";

export default function LeaderboardScreen() {
  const [users, setUsers] = useState<any[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLeaderboard = useCallback(async (pageNum: number, searchQuery: string) => {
    setLoading(true);
    setError(null);
    try {
      const url = `${BACKEND_URL}/leaderboard?page=${pageNum}&limit=50&q=${encodeURIComponent(searchQuery)}`;
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }
      const json = await res.json();

      setUsers(json.users || []);
      setTotalPages(Math.ceil((json.total || 0) / 50));
    } catch (err: any) {
      console.error("Failed to fetch leaderboard", err);
      setError(err.message || "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial fetch and periodic update
  useEffect(() => {
    fetchLeaderboard(page, query);
    const interval = setInterval(() => fetchLeaderboard(page, query), 5000);
    return () => clearInterval(interval);
  }, [page, query, fetchLeaderboard]);

  const handleSearch = (text: string) => {
    setQuery(text);
    setPage(1);
  };

  const goToPage = (p: number) => {
    if (p >= 1 && p <= totalPages) {
      setPage(p);
    }
  };

  // Determine podium and list data
  // Only show podium if page 1 and no active search (or maybe just page 1)
  const showPodium = page === 1 && !query && users.length >= 3;
  // Reorder for 2 - 1 - 3 layout: [Silver, Gold, Bronze]
  const podiumData = showPodium ? [users[1], users[0], users[2]] : [];
  const listUsers = users;

  const getBorderColor = (rank: number) => {
    if (rank === 1) return COLORS.gold;
    if (rank === 2) return COLORS.silver;
    if (rank === 3) return COLORS.bronze;
    return COLORS.primary;
  };

  return (
    <View style={styles.container}>
      <ScrollView contentContainerStyle={styles.scrollContent}>

        {/* Header Section */}
        <View style={styles.headerTitleContainer}>
          <View style={styles.logoContainer}>
            <Image source={{ uri: LOGO_URI }} style={styles.logoImage} resizeMode="contain" />
            <Text style={styles.brandName}>MATIKS</Text>
          </View>
          <View style={styles.pageTitleContainer}>
            <Text style={styles.pageTitle}>LEADERBOARD</Text>
          </View>
        </View>

        {/* Podium Section */}
        {showPodium && (
          <View style={styles.podiumContainer}>
            {podiumData.map((u, index) => {
              if (!u) return null; // Safety check
              const isWinner = index === 1; // Center item is winner (index 1 of our reordered array is actually rank 1)
              // Mapping the visual index back to actual rank for colors
              let rank = 1;
              if (index === 0) rank = 2;
              if (index === 2) rank = 3;

              return (
                <View key={u.username} style={[styles.podiumItem, isWinner && styles.winnerPodiumItem]}>
                  <View style={[styles.avatarContainer, { borderColor: getBorderColor(rank) }]}>
                    <Image
                      source={{ uri: PODIUM_IMAGES[rank] || `https://i.pravatar.cc/150?img=${rank + 10}` }}
                      style={styles.avatar}
                    />
                    <View style={styles.rankBadge}>
                      <Text style={styles.rankBadgeText}>{rank}</Text>
                    </View>
                  </View>
                  <Text style={styles.podiumName} numberOfLines={1}>{u.username}</Text>
                  <Text style={styles.podiumPoints}>{u.rating} pts</Text>
                </View>
              );
            })}
          </View>
        )}

        {/* Search Bar - Moved */}
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Search player..."
            placeholderTextColor="#888"
            value={query}
            onChangeText={handleSearch}
          />
        </View>

        {/* List Header */}
        <View style={styles.tableHeader}>
          <Text style={[styles.hText, { flex: 0.8, textAlign: 'center' }]}>Rank</Text>
          <Text style={[styles.hText, { flex: 2, textAlign: 'left', paddingLeft: 10 }]}>Player</Text>
          <Text style={[styles.hText, { flex: 1.2, textAlign: 'right', paddingRight: 10 }]}>Points</Text>
        </View>

        {/* Loading & Error */}
        {loading && <ActivityIndicator size="large" color={COLORS.primary} style={{ margin: 20 }} />}
        {error && <Text style={styles.errorText}>Error: {error}</Text>}

        {/* List Items */}
        {listUsers.map((u) => (
          <View key={`${u.username}-${u.rank}`} style={styles.tableRow}>
            <View style={styles.rankContainer}>
              <Text style={styles.rankText}>#{u.rank}</Text>
            </View>
            <Text style={styles.checkName} numberOfLines={1}>{u.username}</Text>
            <Text style={styles.checkPoints}>{u.rating}</Text>
          </View>
        ))}

        {/* Pagination */}
        <View style={styles.pagination}>
          <TouchableOpacity
            disabled={page <= 1}
            onPress={() => goToPage(page - 1)}
            style={[styles.pageBtn, page <= 1 && styles.disabledBtn]}
          >
            <Text style={styles.pageBtnText}>Previous</Text>
          </TouchableOpacity>

          <Text style={styles.pageInfo}>
            {page} / {totalPages}
          </Text>

          <TouchableOpacity
            disabled={page >= totalPages}
            onPress={() => goToPage(page + 1)}
            style={[styles.pageBtn, page >= totalPages && styles.disabledBtn]}
          >
            <Text style={styles.pageBtnText}>Next</Text>
          </TouchableOpacity>
        </View>

      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: COLORS.background,
  },
  scrollContent: {
    paddingBottom: 60,
    paddingHorizontal: 10,
  },
  searchContainer: {
    marginVertical: 24,
    marginHorizontal: 16, // Adjusted Spacing
  },
  searchInput: {
    backgroundColor: COLORS.glass,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
    color: COLORS.text,
    paddingVertical: 14,
    paddingHorizontal: 20,
    borderRadius: 16,
    fontSize: 18,
  },

  // Header
  headerTitleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 30, // Increased margin
    paddingHorizontal: 10,
    position: 'relative', // Set to relative to allow absolute centering of title if needed
    height: 60, // set height to align items
  },
  logoContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    position: 'absolute',
    left: 10,
    zIndex: 10,
  },
  logoImage: {
    width: 44,
    height: 44,
    marginRight: 10,
  },
  brandName: {
    color: COLORS.primary,
    fontSize: 26, // Bigger
    fontWeight: 'bold',
    letterSpacing: 2,
  },
  pageTitleContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    // Removed background/border logic if it was looking 'disabled'
    // But keeping it if they asked "dont disable border", I'll make it explicit
    borderWidth: 0, // Removed border
    // borderColor: COLORS.glassBorder, // Removed border color
    backgroundColor: 'transparent', // Made background transparent if desired, or keep it depending on look. User said "remove border". I'll default to just removing border properties, effectively making it just text centering.
    paddingVertical: 12,
    marginHorizontal: 120, // Push it in so it doesn't overlap logo
    borderRadius: 12,
  },
  pageTitle: {
    fontSize: 32, // Much Bigger
    fontWeight: '900',
    color: COLORS.text,
    letterSpacing: 3,
    textTransform: 'uppercase',
  },

  // Podium
  podiumContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'flex-end',
    marginBottom: 40,
    marginTop: 10,
    height: 220,
  },
  podiumItem: {
    alignItems: 'center',
    width: 200, // Slightly smaller width
    marginHorizontal: 30, // More spacing
  },
  winnerPodiumItem: {
    marginBottom: 30,
    transform: [{ scale: 1.4 }], // Significantly bigger
    zIndex: 1,
  },
  avatarContainer: {
    borderWidth: 3,
    borderRadius: 60,
    padding: 3,
    marginBottom: 12,
    position: 'relative',
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
  },
  rankBadge: {
    position: 'absolute',
    bottom: -12,
    alignSelf: 'center',
    backgroundColor: COLORS.background,
    borderRadius: 12,
    paddingVertical: 4,
    paddingHorizontal: 12,
    borderWidth: 1,
    borderColor: COLORS.glassBorder,
    minWidth: 30,
    alignItems: 'center',
  },
  rankBadgeText: {
    color: COLORS.text,
    fontWeight: 'bold',
    fontSize: 14,
  },
  podiumName: {
    color: COLORS.text,
    fontWeight: 'bold',
    fontSize: 16, // Bigger
    marginBottom: 4,
    textAlign: 'center',
  },
  podiumPoints: {
    color: COLORS.primary,
    fontSize: 14, // Bigger
    fontWeight: '600',
  },

  // List Headers
  tableHeader: {
    flexDirection: 'row',
    paddingVertical: 16,
    borderBottomWidth: 1.5, // Thicker
    borderColor: COLORS.glassBorder,
    marginBottom: 12,
    marginHorizontal: 16, // Adjusted Spacing
  },
  hText: {
    color: COLORS.primary,
    fontWeight: '800',
    textTransform: 'uppercase',
    fontSize: 15, // Bigger
    letterSpacing: 1.5,
  },

  // List Rows
  tableRow: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.glass,
    marginBottom: 10,
    paddingVertical: 18, // Bigger touch area
    paddingHorizontal: 16,
    borderRadius: 16,
    borderWidth: 1.5, // Thicker border
    borderColor: COLORS.glassBorder,
    marginHorizontal: 16, // Adjusted Spacing
  },
  rankContainer: {
    flex: 0.5,
    alignItems: 'center',
    justifyContent: 'center',
  },
  rankText: {
    color: COLORS.text,
    fontWeight: 'bold',
    fontSize: 20, // Bigger
    opacity: 1,
  },
  checkName: {
    flex: 2,
    color: COLORS.text,
    fontWeight: '700',
    fontSize: 18, // Bigger
    textAlign: 'left',
    paddingLeft: 10,
  },
  checkPoints: {
    flex: 1.2,
    color: COLORS.primary,
    fontWeight: 'bold',
    fontSize: 18, // Bigger
    textAlign: 'right',
    paddingRight: 10,
  },

  // Pagination
  pagination: {
    flexDirection: "row",
    justifyContent: "space-between", // Spread out
    alignItems: "center",
    marginTop: 30,
    width: '100%',
    paddingHorizontal: 20,
  },
  pageBtn: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    backgroundColor: COLORS.glass,
    borderRadius: 12,
    borderWidth: 2, // Thicker
    borderColor: COLORS.primary,
  },
  disabledBtn: {
    borderColor: '#333',
    opacity: 0.5,
  },
  pageBtnText: {
    color: COLORS.primary,
    fontWeight: "bold",
    fontSize: 14,
    textTransform: 'uppercase',
  },
  pageInfo: {
    color: COLORS.text,
    fontSize: 16,
    fontWeight: 'bold',
  },
  errorText: {
    color: 'red',
    textAlign: 'center',
    marginBottom: 10,
  }
});
