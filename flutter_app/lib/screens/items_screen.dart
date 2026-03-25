import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_app/models/item.dart';
import 'package:flutter_app/providers/items_provider.dart';

/// Main items list screen with inline create/edit/delete.
class ItemsScreen extends ConsumerWidget {
  const ItemsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(itemsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Items'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(itemsProvider.notifier).refresh(),
          ),
        ],
      ),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
        data: (res) => res.data.isEmpty
            ? const Center(child: Text('No items yet. Create one!'))
            : ListView.separated(
                itemCount: res.data.length,
                separatorBuilder: (context, index) => const Divider(height: 1),
                itemBuilder: (context, i) => _ItemTile(item: res.data[i]),
              ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showItemDialog(context, ref),
        child: const Icon(Icons.add),
      ),
    );
  }

  Future<void> _showItemDialog(
    BuildContext context,
    WidgetRef ref, [
    Item? item,
  ]) async {
    final nameCtrl = TextEditingController(text: item?.name ?? '');
    final descCtrl = TextEditingController(text: item?.description ?? '');

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(item == null ? 'Create Item' : 'Edit Item'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameCtrl,
              decoration: const InputDecoration(labelText: 'Name *'),
              autofocus: true,
            ),
            const SizedBox(height: 8),
            TextField(
              controller: descCtrl,
              decoration: const InputDecoration(labelText: 'Description'),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Save'),
          ),
        ],
      ),
    );

    if (confirmed != true || nameCtrl.text.trim().isEmpty) return;

    final notifier = ref.read(itemsProvider.notifier);
    try {
      if (item == null) {
        await notifier.create(
          name: nameCtrl.text.trim(),
          description: descCtrl.text.trim(),
        );
      } else {
        await notifier.updateItem(
          item.id,
          name: nameCtrl.text.trim(),
          description: descCtrl.text.trim(),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(e.toString())));
      }
    }
  }
}

class _ItemTile extends ConsumerWidget {
  final Item item;
  const _ItemTile({required this.item});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ListTile(
      title: Text(item.name),
      subtitle: item.description.isNotEmpty ? Text(item.description) : null,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            onPressed: () => _ItemsScreenState.showEdit(context, ref, item),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            color: Colors.red,
            onPressed: () async {
              final ok = await showDialog<bool>(
                context: context,
                builder: (_) => AlertDialog(
                  title: const Text('Delete item?'),
                  content: Text(
                    'Delete "${item.name}"? This cannot be undone.',
                  ),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(context, false),
                      child: const Text('Cancel'),
                    ),
                    FilledButton(
                      style: FilledButton.styleFrom(
                        backgroundColor: Colors.red,
                      ),
                      onPressed: () => Navigator.pop(context, true),
                      child: const Text('Delete'),
                    ),
                  ],
                ),
              );
              if (ok == true && context.mounted) {
                try {
                  await ref.read(itemsProvider.notifier).delete(item.id);
                } catch (e) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(
                      context,
                    ).showSnackBar(SnackBar(content: Text(e.toString())));
                  }
                }
              }
            },
          ),
        ],
      ),
    );
  }
}

// Static helper so _ItemTile can trigger the edit dialog.
abstract class _ItemsScreenState {
  static void showEdit(BuildContext context, WidgetRef ref, Item item) {
    // Re-use ItemsScreen's dialog logic by instantiating and calling.
    const ItemsScreen()._showItemDialog(context, ref, item);
  }
}
