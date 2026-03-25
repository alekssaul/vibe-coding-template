import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/items_provider.dart';
import '../models/item.dart';
import '../core/widgets/error_state_widget.dart';

/// Main items list screen with inline create/edit/delete.
class ItemsScreen extends ConsumerWidget {
  const ItemsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final itemsState = ref.watch(itemsProvider);

    void refreshItems() {
      ref.invalidate(itemsProvider);
    }

    void deleteItem(int id) async {
      final confirm = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Delete Item'),
          content: const Text('Are you sure you want to delete this item?'),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error,
                foregroundColor: Theme.of(context).colorScheme.onError,
              ),
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Delete'),
            ),
          ],
        ),
      );

      if (confirm == true) {
        await ref.read(itemsProvider.notifier).deleteItem(id);
      }
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Items'),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: refreshItems),
        ],
      ),
      body: itemsState.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => ErrorStateWidget(error: err, onRetry: refreshItems),
        data: (items) => items.isEmpty
            ? const Center(child: Text('No items yet. Create one!'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (context, index) => const Divider(height: 1),
                itemBuilder: (context, i) => _ItemTile(
                  item: items[i],
                  onEdit: (item) => _showItemDialog(context, ref, item),
                  onDelete: deleteItem,
                ),
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
    if (item == null) {
      await notifier.createItem(nameCtrl.text.trim(), descCtrl.text.trim());
    } else {
      await notifier.updateItem(
        item.id,
        nameCtrl.text.trim(),
        descCtrl.text.trim(),
      );
    }
  }
}

class _ItemTile extends ConsumerWidget {
  final Item item;
  final void Function(Item item) onEdit;
  final void Function(int id) onDelete;

  const _ItemTile({
    required this.item,
    required this.onEdit,
    required this.onDelete,
  });

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
            onPressed: () => onEdit(item),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            color: Theme.of(context).colorScheme.error,
            onPressed: () => onDelete(item.id),
          ),
        ],
      ),
    );
  }
}
