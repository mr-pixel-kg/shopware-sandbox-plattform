<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDateTime } from "@/lib/utils";
import type { AuditLogRecord } from "@/types/api";

defineProps<{
  logs: AuditLogRecord[];
}>();
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between space-y-0">
      <div class="space-y-1">
        <CardTitle>Audit log</CardTitle>
        <CardDescription>Recent backend actions for troubleshooting and traceability.</CardDescription>
      </div>
      <Badge variant="secondary">{{ logs.length }} entries</Badge>
    </CardHeader>

    <CardContent>
      <div class="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>User</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Details</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="entry in logs" :key="entry.id">
              <TableCell>{{ formatDateTime(entry.createdAt) }}</TableCell>
              <TableCell class="font-medium">{{ entry.action }}</TableCell>
              <TableCell>{{ entry.userId || "Guest" }}</TableCell>
              <TableCell>{{ entry.ipAddress || "-" }}</TableCell>
              <TableCell class="max-w-lg text-xs text-muted-foreground">
                <pre class="whitespace-pre-wrap break-all">{{ JSON.stringify(entry.details ?? {}, null, 2) }}</pre>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>
</template>
